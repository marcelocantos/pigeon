// pigeon-demo runs the entire pigeon stack — relay, backend, and a
// configurable number of clients — in a single process behind an HTTP
// control server, then opens a web UI showing the live message flow on
// every channel for every client, all multiplexed onto the one backend.
//
// Architecture
//
//	┌──────────┐                              ┌─────────┐
//	│ client-1 │ ──┐                       ┌─►│         │
//	│  Session │ ──┤        ┌─────────┐    │  │ backend │
//	└──────────┘   ├──QUIC─►│  relay  │◄───┤  │ Session │
//	┌──────────┐   │        │ (mux)   │    │  │  (one)  │
//	│ client-N │ ──┘        └─────────┘    └──│         │
//	│  Session │                              └────┬────┘
//	└──────────┘                              events │
//	     │ events                                    │
//	     ▼                                           ▼
//	         ┌───────────────────────────────┐
//	         │  control HTTP server (this)   │
//	         │  /events (SSE) /client/{n}/*  │
//	         └───────────────┬───────────────┘
//	                         │ HTTP/SSE
//	                         ▼
//	                 ┌──────────────┐
//	                 │  browser UI  │
//	                 └──────────────┘
//
// The pairing ceremony runs in-process at startup with auto-confirm —
// no QR codes, no user prompts.
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
	"strconv"
	"strings"
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
// to the browser; the UI uses Pane / ClientID / Dir to decide where
// and how to render it.
//
//	Pane:     "client" | "backend" | "system"
//	ClientID: "client-1" / "client-2" / … on every client and backend
//	          event so the UI can group and colour by client.
//	Dir:      "out" | "in" | "info"
//	Channel:  "chat" | "control" | "ping" | "metric" | ""
type event struct {
	Time     time.Time `json:"time"`
	Pane     string    `json:"pane"`
	ClientID string    `json:"clientId,omitempty"`
	Dir      string    `json:"dir"`
	Channel  string    `json:"channel"`
	Text     string    `json:"text"`
}

// bus fans out events to every connected SSE subscriber.
type bus struct {
	mu      sync.Mutex
	subs    map[chan event]struct{}
	history []event
	maxHist int
}

func newBus() *bus {
	return &bus{subs: map[chan event]struct{}{}, maxHist: 400}
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
		}
	}
}

func (b *bus) subscribe() (chan event, []event, func()) {
	c := make(chan event, 128)
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

// clientHarness is one connected client with its four channels and
// closure-based send methods that publish "out" events on the bus.
type clientHarness struct {
	id          string
	instance    string
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

func main() {
	httpPort := flag.Int("http-port", 7000, "control HTTP port (web UI lives here)")
	openBrowser := flag.Bool("open", true, "open the demo UI in the default browser at startup")
	numClients := flag.Int("clients", 3, "how many clients to spawn (1..16)")
	flag.Parse()

	if *numClients < 1 || *numClients > 16 {
		fmt.Fprintln(os.Stderr, "--clients must be between 1 and 16")
		os.Exit(2)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	b := newBus()

	relayURL, err := startRelay(ctx, b)
	if err != nil {
		slog.Error("start relay", "err", err)
		os.Exit(1)
	}
	b.publish(event{Pane: "system", Dir: "info", Text: "relay up at " + relayURL})

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

	// Pair every client against the same backend up-front. This produces
	// one PairingRecord-per-client on each side (the backend's lookup
	// table; the client's own connect record).
	type clientSeed struct {
		id        string
		identity  crypto.Identity
		clientRec *crypto.PairingRecord
	}
	seeds := make([]clientSeed, 0, *numClients)
	pairings := make(map[string]*crypto.PairingRecord, *numClients)

	for i := 1; i <= *numClients; i++ {
		cid, err := crypto.NewFileIdentity(fmt.Sprintf("%s/client-%d-id.json", tmp, i))
		if err != nil {
			slog.Error("client identity", "i", i, "err", err)
			os.Exit(1)
		}
		brec, crec, err := autoPair(ctx, relayURL, bid, cid, b)
		if err != nil {
			slog.Error("auto-pair", "i", i, "err", err)
			os.Exit(1)
		}
		pairings[cid.InstanceID()] = brec
		seeds = append(seeds, clientSeed{
			id:        fmt.Sprintf("client-%d", i),
			identity:  cid,
			clientRec: crec,
		})
		b.publish(event{Pane: "system", Dir: "info", Text: fmt.Sprintf("paired client-%d (%s)", i, cid.InstanceID())})
	}

	// Start the backend Listener with the full pairings table.
	backendReady := make(chan struct{})
	// idLookup maps a client InstanceID to its friendly demo name so
	// backend events can attribute traffic to the right client tab.
	idLookup := make(map[string]string, len(seeds))
	for _, s := range seeds {
		idLookup[s.identity.InstanceID()] = s.id
	}
	go runBackend(ctx, relayURL, bid, pairings, idLookup, b, backendReady)

	select {
	case <-backendReady:
	case <-ctx.Done():
		return
	}

	// Connect every client.
	clients := make([]*clientHarness, 0, len(seeds))
	for _, s := range seeds {
		c, err := startClient(ctx, relayURL, s.id, s.identity, bid.InstanceID(), s.clientRec, b)
		if err != nil {
			slog.Error("start client", "id", s.id, "err", err)
			os.Exit(1)
		}
		clients = append(clients, c)
		defer c.close()
	}

	mux := http.NewServeMux()

	sub, _ := fs.Sub(webFS, "web")
	mux.Handle("/", http.FileServer(http.FS(sub)))

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
		write := func(ev event) {
			data, _ := json.Marshal(ev)
			fmt.Fprintf(w, "data: %s\n\n", data)
			fl.Flush()
		}
		for _, ev := range hist {
			write(ev)
		}
		for {
			select {
			case ev, ok := <-ch:
				if !ok {
					return
				}
				write(ev)
			case <-r.Context().Done():
				return
			}
		}
	})

	mux.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		type clientState struct {
			ID       string `json:"id"`
			Instance string `json:"instance"`
		}
		cs := make([]clientState, 0, len(clients))
		for _, c := range clients {
			cs = append(cs, clientState{ID: c.id, Instance: c.instance})
		}
		st := map[string]any{
			"relay":            relayURL,
			"backendInstance":  bid.InstanceID(),
			"clients":          cs,
			"streamChannels":   []string{"chat", "control"},
			"datagramChannels": dgChannels,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(st)
	})

	// /client/{n}/{action} — find harness, dispatch.
	mux.HandleFunc("/client/", func(w http.ResponseWriter, r *http.Request) {
		// path is "/client/<n>/<action>"
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/client/"), "/")
		if len(parts) != 2 {
			http.Error(w, "bad path", http.StatusNotFound)
			return
		}
		idx, err := strconv.Atoi(parts[0])
		if err != nil || idx < 1 || idx > len(clients) {
			http.Error(w, "bad client index", http.StatusNotFound)
			return
		}
		c := clients[idx-1]
		txt := readText(r)
		var send func() error
		switch parts[1] {
		case "chat":
			send = func() error { return c.sendChat(txt) }
		case "control":
			send = func() error { return c.sendControl(txt) }
		case "ping":
			send = func() error { return c.sendPing([]byte(txt)) }
		case "metric":
			send = func() error { return c.sendMetric([]byte("get")) }
		default:
			http.Error(w, "bad action", http.StatusNotFound)
			return
		}
		if err := send(); err != nil {
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
	slog.Info("demo ready", "url", url, "clients", len(clients))
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintf(os.Stderr, "  ┌─ pigeon demo ready (%d clients) ─────────────────────\n", len(clients))
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

func startRelay(ctx context.Context, b *bus) (string, error) {
	cert, err := selfSignedCert()
	if err != nil {
		return "", fmt.Errorf("cert: %w", err)
	}
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	wtSrv, err := pigeon.NewWebTransportServer("127.0.0.1:0", tlsCfg, "")
	if err != nil {
		return "", fmt.Errorf("wt server: %w", err)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("resolve: %w", err)
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return "", fmt.Errorf("listen udp: %w", err)
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

	time.Sleep(100 * time.Millisecond)
	return fmt.Sprintf("https://127.0.0.1:%d", port), nil
}

// autoPair runs the pairing ceremony for one (backend, client) pair
// and returns (backendRecord, clientRecord). Auto-confirms on both
// sides as soon as the codes match.
func autoPair(ctx context.Context, relayURL string, bid, cid crypto.Identity, _ *bus) (*crypto.PairingRecord, *crypto.PairingRecord, error) {
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

		_, err = cer.Code(ctx)
		if err != nil {
			accCh <- accResult{err: fmt.Errorf("backend code: %w", err)}
			return
		}
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
	if _, err := icer.Code(ctx); err != nil {
		return nil, nil, fmt.Errorf("client code: %w", err)
	}
	clientRec, err := icer.Confirm(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("client confirm: %w", err)
	}

	res := <-accCh
	if res.err != nil {
		return nil, nil, fmt.Errorf("backend confirm: %w", res.err)
	}
	return res.rec, clientRec, nil
}

func runBackend(
	ctx context.Context,
	relayURL string,
	identity crypto.Identity,
	pairings map[string]*crypto.PairingRecord,
	idLookup map[string]string,
	b *bus,
	ready chan<- struct{},
) {
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
		friendly := idLookup[sess.PeerID()]
		if friendly == "" {
			friendly = sess.PeerID()
		}
		go serveBackendSession(ctx, sess, friendly, b)
	}
}

func serveBackendSession(ctx context.Context, sess *pigeon.Session, clientID string, b *bus) {
	defer sess.Close()
	b.publish(event{Pane: "backend", ClientID: clientID, Dir: "info", Text: clientID + " connected"})
	defer b.publish(event{Pane: "backend", ClientID: clientID, Dir: "info", Text: clientID + " disconnected"})

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
			b.publish(event{Pane: "backend", ClientID: clientID, Dir: "in", Channel: "chat", Text: string(msg)})
			reply := append([]byte("echo: "), msg...)
			if err := chat.Send(reply); err != nil {
				return
			}
			b.publish(event{Pane: "backend", ClientID: clientID, Dir: "out", Channel: "chat", Text: string(reply)})
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
			b.publish(event{Pane: "backend", ClientID: clientID, Dir: "in", Channel: "control", Text: string(msg)})
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
			b.publish(event{Pane: "backend", ClientID: clientID, Dir: "out", Channel: "control", Text: string(reply)})
		}
	}()

	go func() {
		ping := sess.Datagram("ping")
		for {
			p, err := ping.Recv(ctx)
			if err != nil {
				return
			}
			b.publish(event{Pane: "backend", ClientID: clientID, Dir: "in", Channel: "ping", Text: string(p)})
			reply := append([]byte("pong:"), p...)
			if err := ping.Send(reply); err != nil {
				return
			}
			b.publish(event{Pane: "backend", ClientID: clientID, Dir: "out", Channel: "ping", Text: string(reply)})
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
			b.publish(event{Pane: "backend", ClientID: clientID, Dir: "in", Channel: "metric", Text: string(req)})
			echoMu.Lock()
			reply := fmt.Appendf(nil, "echoes=%d", echoes)
			echoMu.Unlock()
			if err := metric.Send(reply); err != nil {
				return
			}
			b.publish(event{Pane: "backend", ClientID: clientID, Dir: "out", Channel: "metric", Text: string(reply)})
		}
	}()

	<-ctx.Done()
}

func startClient(ctx context.Context, relayURL, demoID string, identity crypto.Identity, backendInstance string, crec *crypto.PairingRecord, b *bus) (*clientHarness, error) {
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
	b.publish(event{Pane: "client", ClientID: demoID, Dir: "info", Text: "connected to " + backendInstance})

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

	c := &clientHarness{id: demoID, instance: identity.InstanceID(), sess: sess}

	// Receive pumps: every inbound message tagged with this client's
	// demoID so the UI can route it to the right client tab.
	go func() {
		for {
			msg, err := chat.Recv(ctx)
			if err != nil {
				return
			}
			b.publish(event{Pane: "client", ClientID: demoID, Dir: "in", Channel: "chat", Text: string(msg)})
		}
	}()
	go func() {
		for {
			msg, err := ctrl.Recv(ctx)
			if err != nil {
				return
			}
			b.publish(event{Pane: "client", ClientID: demoID, Dir: "in", Channel: "control", Text: string(msg)})
		}
	}()
	go func() {
		for {
			p, err := ping.Recv(ctx)
			if err != nil {
				return
			}
			b.publish(event{Pane: "client", ClientID: demoID, Dir: "in", Channel: "ping", Text: string(p)})
		}
	}()
	go func() {
		for {
			m, err := metric.Recv(ctx)
			if err != nil {
				return
			}
			b.publish(event{Pane: "client", ClientID: demoID, Dir: "in", Channel: "metric", Text: string(m)})
		}
	}()

	c.sendChat = func(text string) error {
		if err := chat.Send([]byte(text)); err != nil {
			return err
		}
		b.publish(event{Pane: "client", ClientID: demoID, Dir: "out", Channel: "chat", Text: text})
		return nil
	}
	c.sendControl = func(text string) error {
		if err := ctrl.Send([]byte(text)); err != nil {
			return err
		}
		b.publish(event{Pane: "client", ClientID: demoID, Dir: "out", Channel: "control", Text: text})
		return nil
	}
	c.sendPing = func(p []byte) error {
		if err := ping.Send(p); err != nil {
			return err
		}
		b.publish(event{Pane: "client", ClientID: demoID, Dir: "out", Channel: "ping", Text: string(p)})
		return nil
	}
	c.sendMetric = func(p []byte) error {
		if err := metric.Send(p); err != nil {
			return err
		}
		b.publish(event{Pane: "client", ClientID: demoID, Dir: "out", Channel: "metric", Text: string(p)})
		return nil
	}

	return c, nil
}

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
