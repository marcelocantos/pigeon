// pigeon-demo runs the entire pigeon stack — relay, backend, client —
// in a single process behind an HTTP control server, then opens a web
// UI showing the live message flow on every channel.
//
// Architecture
//
//	┌─────────┐   QUIC   ┌─────────┐   QUIC   ┌─────────┐
//	│ client  │ ───────► │  relay  │ ───────► │ backend │
//	│ Session │ ◄─────── │ (mux)   │ ◄─────── │ Session │
//	└────┬────┘                                └────┬────┘
//	     │ events                              events │
//	     ▼                                            ▼
//	         ┌───────────────────────────────┐
//	         │  control HTTP server (this)   │
//	         │  /events (SSE) /client/*      │
//	         └───────────────┬───────────────┘
//	                         │ HTTP/SSE
//	                         ▼
//	                 ┌──────────────┐
//	                 │  browser UI  │
//	                 └──────────────┘
//
// The pairing ceremony runs in-process at startup with auto-confirm —
// no QR codes, no user prompts. Both peers use the same crypto.Identity
// and PairingRecord types as the public API; this is a real pairing,
// just non-interactive.
//
// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/marcelocantos/pigeon"
	"github.com/marcelocantos/pigeon/crypto"
	"github.com/marcelocantos/pigeon/pairing"
)

//go:embed web/*
var webFS embed.FS

// dgChannels matches what the echo example uses.
var dgChannels = map[string]uint64{
	"ping":   1,
	"metric": 2,
}

// event is a single line in the live message-flow log. Sent over SSE
// to the browser; the UI picks "pane" and "dir" to decide where to
// render it (client pane, backend pane; arrow direction).
type event struct {
	Time    time.Time `json:"time"`
	Pane    string    `json:"pane"`    // "client" | "backend" | "system"
	Dir     string    `json:"dir"`     // "out" | "in" | "info"
	Channel string    `json:"channel"` // "chat" | "control" | "ping" | "metric" | ""
	Text    string    `json:"text"`
}

// bus fans out events to every connected SSE subscriber.
type bus struct {
	mu       sync.Mutex
	subs     map[chan event]struct{}
	history  []event
	maxHist  int
	seq      uint64
}

func newBus() *bus {
	return &bus{subs: map[chan event]struct{}{}, maxHist: 200}
}

func (b *bus) publish(ev event) {
	if ev.Time.IsZero() {
		ev.Time = time.Now()
	}
	b.mu.Lock()
	b.history = append(b.history, ev)
	if len(b.history) > b.maxHist {
		b.history = b.history[len(b.history)-b.maxHist:]
	}
	subs := make([]chan event, 0, len(b.subs))
	for c := range b.subs {
		subs = append(subs, c)
	}
	b.mu.Unlock()
	for _, c := range subs {
		select {
		case c <- ev:
		default:
			// Slow subscriber — drop rather than block.
		}
	}
}

func (b *bus) subscribe() (chan event, []event, func()) {
	c := make(chan event, 64)
	b.mu.Lock()
	b.subs[c] = struct{}{}
	hist := append([]event(nil), b.history...)
	b.mu.Unlock()
	cancel := func() {
		b.mu.Lock()
		delete(b.subs, c)
		b.mu.Unlock()
		close(c)
	}
	return c, hist, cancel
}

func main() {
	httpPort := flag.Int("http-port", 7000, "control HTTP port (web UI lives here)")
	openBrowser := flag.Bool("open", true, "open the demo UI in the default browser at startup")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	b := newBus()

	// 1. Start relay (raw QUIC on a free port; self-signed cert).
	relayURL, _, err := startRelay(ctx, b)
	if err != nil {
		slog.Error("start relay", "err", err)
		os.Exit(1)
	}
	b.publish(event{Pane: "system", Dir: "info", Text: "relay up at " + relayURL})

	// 2. Generate two identities.
	tmp, err := os.MkdirTemp("", "pigeon-demo-")
	if err != nil {
		slog.Error("tempdir", "err", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmp)
	bid, err := crypto.NewFileIdentity(tmp + "/backend-id.json")
	if err != nil {
		slog.Error("backend identity", "err", err)
		os.Exit(1)
	}
	cid, err := crypto.NewFileIdentity(tmp + "/client-id.json")
	if err != nil {
		slog.Error("client identity", "err", err)
		os.Exit(1)
	}

	// 3. Pair them in-process (auto-confirm).
	brec, crec, err := autoPair(ctx, relayURL, bid, cid, b)
	if err != nil {
		slog.Error("auto-pair", "err", err)
		os.Exit(1)
	}
	b.publish(event{Pane: "system", Dir: "info", Text: "pairing complete (backend ↔ client)"})

	// 4. Start the backend Listener loop.
	backendReady := make(chan struct{})
	go runBackend(ctx, relayURL, bid, cid.InstanceID(), brec, b, backendReady)

	// Wait until the backend is registered (so the client can connect).
	select {
	case <-backendReady:
	case <-ctx.Done():
		return
	}

	// 5. Connect the client.
	cli, err := startClient(ctx, relayURL, cid, bid.InstanceID(), crec, b)
	if err != nil {
		slog.Error("start client", "err", err)
		os.Exit(1)
	}
	defer cli.close()

	// 6. Start the HTTP control server (web UI + SSE + POST endpoints).
	mux := http.NewServeMux()

	// Static web UI.
	sub, _ := fs.Sub(webFS, "web")
	mux.Handle("/", http.FileServer(http.FS(sub)))

	// SSE event stream.
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("X-Accel-Buffering", "no")
		fl, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "no flusher", http.StatusInternalServerError)
			return
		}
		ch, hist, unsub := b.subscribe()
		defer unsub()

		writeEv := func(ev event) {
			data, _ := json.Marshal(ev)
			fmt.Fprintf(w, "data: %s\n\n", data)
			fl.Flush()
		}
		// Replay history first so the UI has context on (re)connect.
		for _, ev := range hist {
			writeEv(ev)
		}
		for {
			select {
			case ev, ok := <-ch:
				if !ok {
					return
				}
				writeEv(ev)
			case <-r.Context().Done():
				return
			}
		}
	})

	// Snapshot of the wired-up state (instance IDs, channels) for the UI.
	mux.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		st := map[string]any{
			"relay":            relayURL,
			"backendInstance":  bid.InstanceID(),
			"clientInstance":   cid.InstanceID(),
			"streamChannels":   []string{"chat", "control"},
			"datagramChannels": dgChannels,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(st)
	})

	// Send-on-channel POST endpoints.
	mux.HandleFunc("/client/chat", func(w http.ResponseWriter, r *http.Request) {
		txt := readText(r)
		if err := cli.sendChat(txt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/client/control", func(w http.ResponseWriter, r *http.Request) {
		txt := readText(r)
		if err := cli.sendControl(txt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/client/ping", func(w http.ResponseWriter, r *http.Request) {
		txt := readText(r)
		if err := cli.sendPing([]byte(txt)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/client/metric", func(w http.ResponseWriter, r *http.Request) {
		if err := cli.sendMetric([]byte("get")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	addr := fmt.Sprintf("127.0.0.1:%d", *httpPort)
	srv := &http.Server{Addr: addr, Handler: mux}
	go func() {
		<-ctx.Done()
		shutCtx, cancelShut := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancelShut()
		_ = srv.Shutdown(shutCtx)
	}()
	url := "http://" + addr
	slog.Info("demo ready", "url", url)
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  ┌─ pigeon demo ready ─────────────────────────────────")
	fmt.Fprintln(os.Stderr, "  │  open: "+url)
	fmt.Fprintln(os.Stderr, "  └─────────────────────────────────────────────────────")
	fmt.Fprintln(os.Stderr, "")

	if *openBrowser {
		go func() {
			time.Sleep(300 * time.Millisecond)
			_ = browseURL(url)
		}()
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("http server", "err", err)
	}
}

func readText(r *http.Request) string {
	defer r.Body.Close()
	var body struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return ""
	}
	return body.Text
}

func browseURL(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}
	return nil
}

// startRelay brings up an in-process raw-QUIC pigeon relay on a free
// UDP port and returns its https://127.0.0.1:PORT URL.
func startRelay(ctx context.Context, b *bus) (string, *pigeon.QUICServer, error) {
	cert, err := selfSignedCert()
	if err != nil {
		return "", nil, fmt.Errorf("cert: %w", err)
	}
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	wtSrv, err := pigeon.NewWebTransportServer("127.0.0.1:0", tlsCfg, "")
	if err != nil {
		return "", nil, fmt.Errorf("wt server: %w", err)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		return "", nil, fmt.Errorf("resolve: %w", err)
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return "", nil, fmt.Errorf("listen udp: %w", err)
	}
	port := udpConn.LocalAddr().(*net.UDPAddr).Port
	qSrv := pigeon.NewQUICServer(fmt.Sprintf("127.0.0.1:%d", port), tlsCfg, "", wtSrv.Hub())
	go func() {
		if err := qSrv.ServeWithTLS(udpConn, tlsCfg); err != nil && ctx.Err() == nil {
			b.publish(event{Pane: "system", Dir: "info", Text: "relay stopped: " + err.Error()})
		}
	}()

	go func() {
		<-ctx.Done()
		_ = qSrv.Close()
		_ = wtSrv.Close()
	}()

	// quic-go binds synchronously inside Listen, but allow a tick for
	// the listener goroutine to schedule.
	time.Sleep(100 * time.Millisecond)

	return fmt.Sprintf("https://127.0.0.1:%d", port), qSrv, nil
}

// autoPair runs the pairing ceremony on both sides in this process and
// returns the resulting (backendRecord, clientRecord) pair, with both
// sides auto-confirming once the codes match.
func autoPair(ctx context.Context, relayURL string, bid, cid crypto.Identity, b *bus) (*crypto.PairingRecord, *crypto.PairingRecord, error) {
	pairer, err := pairing.Register(ctx, &pairing.Args{Relay: relayURL, Identity: bid})
	if err != nil {
		return nil, nil, fmt.Errorf("pairing.Register: %w", err)
	}
	defer pairer.Close()

	type accResult struct {
		rec *crypto.PairingRecord
		err error
	}
	accCh := make(chan accResult, 1)
	tokenCh := make(chan string, 1)

	go func() {
		cer, err := pairer.Accept(ctx)
		if err != nil {
			accCh <- accResult{err: fmt.Errorf("pairer.Accept: %w", err)}
			return
		}
		defer cer.Close()
		tokenCh <- cer.Token

		bcode, err := cer.Code(ctx)
		if err != nil {
			accCh <- accResult{err: fmt.Errorf("backend code: %w", err)}
			return
		}
		b.publish(event{Pane: "system", Dir: "info", Text: "pairing code (backend): " + bcode})
		rec, err := cer.Confirm(ctx)
		accCh <- accResult{rec: rec, err: err}
	}()

	var token string
	select {
	case token = <-tokenCh:
	case <-time.After(10 * time.Second):
		return nil, nil, fmt.Errorf("timed out waiting for backend pairing token")
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	}

	icer, err := pairing.Initiate(ctx, &pairing.InitiateArgs{Rendezvous: token, Identity: cid})
	if err != nil {
		return nil, nil, fmt.Errorf("pairing.Initiate: %w", err)
	}
	defer icer.Close()
	icode, err := icer.Code(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("client code: %w", err)
	}
	b.publish(event{Pane: "system", Dir: "info", Text: "pairing code (client): " + icode})
	clientRec, err := icer.Confirm(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("client confirm: %w", err)
	}

	res := <-accCh
	if res.err != nil {
		return nil, nil, fmt.Errorf("backend confirm: %w", res.err)
	}
	if icode != "" && res.rec != nil {
		// Codes were derived independently and must match; if they don't,
		// the ceremony would have aborted, so this is just an assertion.
	}
	return res.rec, clientRec, nil
}

// runBackend registers the backend on the relay and serves accepted
// Sessions on chat/control/ping/metric. Each message in or out emits
// an event on the bus so the web UI can render it.
func runBackend(ctx context.Context, relayURL string, identity crypto.Identity, expectedClientID string, brec *crypto.PairingRecord, b *bus, ready chan<- struct{}) {
	pairings := map[string]*crypto.PairingRecord{expectedClientID: brec}
	tlsCfg := &tls.Config{InsecureSkipVerify: true}

	lis, id, err := pigeon.Register(ctx, &pigeon.RegisterArgs{
		Identity: identity,
		Pairing: func(clientID string) (*crypto.PairingRecord, error) {
			rec, ok := pairings[clientID]
			if !ok {
				return nil, fmt.Errorf("unknown client %q", clientID)
			}
			return rec, nil
		},
		Relay:     relayURL,
		TLS:       tlsCfg,
		Datagrams: dgChannels,
	})
	if err != nil {
		b.publish(event{Pane: "backend", Dir: "info", Text: "register failed: " + err.Error()})
		return
	}
	defer lis.Close()
	b.publish(event{Pane: "backend", Dir: "info", Text: "registered as " + id})
	close(ready)

	for {
		sess, err := lis.Accept(ctx)
		if err != nil {
			if ctx.Err() == nil {
				b.publish(event{Pane: "backend", Dir: "info", Text: "accept error: " + err.Error()})
			}
			return
		}
		go serveBackendSession(ctx, sess, b)
	}
}

func serveBackendSession(ctx context.Context, sess *pigeon.Session, b *bus) {
	defer sess.Close()
	b.publish(event{Pane: "backend", Dir: "info", Text: "client connected: " + sess.PeerID()})
	defer b.publish(event{Pane: "backend", Dir: "info", Text: "client disconnected"})

	var echoes int
	var echoMu sync.Mutex

	go func() {
		chat, err := sess.AcceptStream(ctx, "chat")
		if err != nil {
			return
		}
		for {
			msg, err := chat.Recv(ctx)
			if err != nil {
				return
			}
			b.publish(event{Pane: "backend", Dir: "in", Channel: "chat", Text: string(msg)})
			reply := append([]byte("echo: "), msg...)
			if err := chat.Send(reply); err != nil {
				return
			}
			b.publish(event{Pane: "backend", Dir: "out", Channel: "chat", Text: string(reply)})
			echoMu.Lock()
			echoes++
			echoMu.Unlock()
		}
	}()

	go func() {
		ctrl, err := sess.AcceptStream(ctx, "control")
		if err != nil {
			return
		}
		for {
			msg, err := ctrl.Recv(ctx)
			if err != nil {
				return
			}
			b.publish(event{Pane: "backend", Dir: "in", Channel: "control", Text: string(msg)})
			var reply []byte
			switch string(msg) {
			case "stats":
				echoMu.Lock()
				reply = fmt.Appendf(nil, "echoes=%d", echoes)
				echoMu.Unlock()
			default:
				reply = fmt.Appendf(nil, "unknown command: %s", msg)
			}
			if err := ctrl.Send(reply); err != nil {
				return
			}
			b.publish(event{Pane: "backend", Dir: "out", Channel: "control", Text: string(reply)})
		}
	}()

	go func() {
		ping := sess.Datagram("ping")
		for {
			p, err := ping.Recv(ctx)
			if err != nil {
				return
			}
			b.publish(event{Pane: "backend", Dir: "in", Channel: "ping", Text: string(p)})
			reply := append([]byte("pong:"), p...)
			if err := ping.Send(reply); err != nil {
				return
			}
			b.publish(event{Pane: "backend", Dir: "out", Channel: "ping", Text: string(reply)})
			echoMu.Lock()
			echoes++
			echoMu.Unlock()
		}
	}()

	go func() {
		metric := sess.Datagram("metric")
		for {
			req, err := metric.Recv(ctx)
			if err != nil {
				return
			}
			b.publish(event{Pane: "backend", Dir: "in", Channel: "metric", Text: string(req)})
			echoMu.Lock()
			reply := fmt.Appendf(nil, "echoes=%d", echoes)
			echoMu.Unlock()
			if err := metric.Send(reply); err != nil {
				return
			}
			b.publish(event{Pane: "backend", Dir: "out", Channel: "metric", Text: string(reply)})
		}
	}()

	<-ctx.Done()
}

// clientHarness wraps a connected pigeon.Session with the four channel
// objects pre-opened, plus an event-publishing receive loop on each.
// Send methods are function-valued fields so startClient can install
// closures that publish "out" events alongside the actual transmit.
type clientHarness struct {
	sess        *pigeon.Session
	sendChat    func(text string) error
	sendControl func(text string) error
	sendPing    func(p []byte) error
	sendMetric  func(p []byte) error
}

func (c *clientHarness) close() {
	if c.sess != nil {
		_ = c.sess.Close()
	}
}

func startClient(ctx context.Context, relayURL string, identity crypto.Identity, backendInstance string, crec *crypto.PairingRecord, b *bus) (*clientHarness, error) {
	tlsCfg := &tls.Config{InsecureSkipVerify: true}
	sess, err := pigeon.Connect(ctx, &pigeon.ConnectArgs{
		InstanceID: backendInstance,
		Record:     crec,
		Identity:   identity,
		Relay:      relayURL,
		TLS:        tlsCfg,
		Datagrams:  dgChannels,
	})
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	b.publish(event{Pane: "client", Dir: "info", Text: "connected to " + backendInstance})

	chat, err := sess.OpenStream(ctx, "chat")
	if err != nil {
		_ = sess.Close()
		return nil, fmt.Errorf("open chat: %w", err)
	}
	ctrl, err := sess.OpenStream(ctx, "control")
	if err != nil {
		_ = sess.Close()
		return nil, fmt.Errorf("open control: %w", err)
	}
	ping := sess.Datagram("ping")
	metric := sess.Datagram("metric")

	c := &clientHarness{sess: sess}

	// Spawn receive goroutines for each channel.
	go func() {
		for {
			msg, err := chat.Recv(ctx)
			if err != nil {
				return
			}
			b.publish(event{Pane: "client", Dir: "in", Channel: "chat", Text: string(msg)})
		}
	}()
	go func() {
		for {
			msg, err := ctrl.Recv(ctx)
			if err != nil {
				return
			}
			b.publish(event{Pane: "client", Dir: "in", Channel: "control", Text: string(msg)})
		}
	}()
	go func() {
		for {
			p, err := ping.Recv(ctx)
			if err != nil {
				return
			}
			b.publish(event{Pane: "client", Dir: "in", Channel: "ping", Text: string(p)})
		}
	}()
	go func() {
		for {
			m, err := metric.Recv(ctx)
			if err != nil {
				return
			}
			b.publish(event{Pane: "client", Dir: "in", Channel: "metric", Text: string(m)})
		}
	}()

	// Also publish "out" events for each send. Wrap by replacing the
	// methods with bus-aware variants.
	c.sendChat = func(text string) error {
		if err := chat.Send([]byte(text)); err != nil {
			return err
		}
		b.publish(event{Pane: "client", Dir: "out", Channel: "chat", Text: text})
		return nil
	}
	c.sendControl = func(text string) error {
		if err := ctrl.Send([]byte(text)); err != nil {
			return err
		}
		b.publish(event{Pane: "client", Dir: "out", Channel: "control", Text: text})
		return nil
	}
	c.sendPing = func(p []byte) error {
		if err := ping.Send(p); err != nil {
			return err
		}
		b.publish(event{Pane: "client", Dir: "out", Channel: "ping", Text: string(p)})
		return nil
	}
	c.sendMetric = func(p []byte) error {
		if err := metric.Send(p); err != nil {
			return err
		}
		b.publish(event{Pane: "client", Dir: "out", Channel: "metric", Text: string(p)})
		return nil
	}

	return c, nil
}

// selfSignedCert produces a single-day-validity P-256 self-signed cert
// for 127.0.0.1 — fine for the demo, never use in production.
func selfSignedCert() (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}
	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	notBefore := time.Now().Add(-time.Hour)
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    notBefore,
		NotAfter:     notBefore.Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}, nil
}
