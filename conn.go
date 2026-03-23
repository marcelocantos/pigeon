// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"

	"github.com/marcelocantos/tern/crypto"
	"github.com/marcelocantos/tern/qr"
)

// Internal message types — first byte of encrypted plaintext.
// These are invisible to callers; only application messages are delivered.
const (
	msgApp      byte = 0x00 // application message
	msgLANOffer byte = 0x01 // LAN address exchange
	msgCutover  byte = 0x02 // transport cutover marker
)

// transport wraps a single WebSocket connection.
type transport struct {
	ws      *websocket.Conn
	name    string // "relay", "lan"
	cutover bool   // true after peer's CUTOVER received on this transport
}

// incomingMsg is a message read by a transport reader goroutine.
type incomingMsg struct {
	transportIdx int
	mt           websocket.MessageType
	data         []byte
	err          error
}

// Conn manages communication with a peer through one or more transports.
// In raw mode (before SetChannel), it passes bytes through a single
// transport unchanged. In encrypted mode (after SetChannel), it handles
// encryption, message framing, reordering across transports, LAN
// discovery, and transport cutover — all transparently to the caller.
type Conn struct {
	mu         sync.Mutex
	instanceID string
	relayURL   string

	// Transports. Index 0 is always the relay.
	transports []*transport
	preferred  int // index into transports; send goes here

	// Encryption. Nil means raw mode.
	channel *crypto.Channel

	// Reader infrastructure.
	incoming chan incomingMsg
	ctx      context.Context    // Conn lifecycle context
	cancel   context.CancelFunc // cancels reader goroutines

	// Reorder buffer (encrypted mode only). Maps sequence number
	// to buffered message. Only populated during transport cutover
	// when messages may arrive out of order across transports.
	reorderBuf  map[uint64]reorderEntry
	nextRecvSeq uint64

	// Application message delivery channel. In encrypted mode, the
	// background pump processes control messages and delivers app
	// messages here for Recv to pick up.
	appMessages chan appMsg

	// LAN upgrade state.
	lanAttempted bool
	lanListener  net.Listener
	lanPort      int // port the LAN listener is bound to
}

type appMsg struct {
	mt   websocket.MessageType
	data []byte
	err  error
}

type reorderEntry struct {
	mt   websocket.MessageType
	data []byte
}

func newConn(ws *websocket.Conn, instanceID, relayURL string) *Conn {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Conn{
		instanceID: instanceID,
		relayURL:   relayURL,
		transports: []*transport{{ws: ws, name: "relay"}},
		incoming:   make(chan incomingMsg, 16),
		ctx:        ctx,
		cancel:     cancel,
		reorderBuf: make(map[uint64]reorderEntry),
	}
	go c.readLoop(0, ws)
	return c
}

// readLoop reads from a transport and feeds messages into the incoming channel.
func (c *Conn) readLoop(idx int, ws *websocket.Conn) {
	for {
		mt, data, err := ws.Read(c.ctx)
		select {
		case c.incoming <- incomingMsg{transportIdx: idx, mt: mt, data: data, err: err}:
		case <-c.ctx.Done():
			return
		}
		if err != nil {
			return
		}
	}
}

// InstanceID returns the relay-assigned instance ID.
func (c *Conn) InstanceID() string {
	return c.instanceID
}

// SetChannel enables encrypted mode. After this call:
//   - Send encrypts plaintext and adds internal framing
//   - Recv decrypts and strips framing, delivering only application messages
//   - LAN discovery begins automatically in the background
//   - Transport cutover is handled transparently
func (c *Conn) SetChannel(ch *crypto.Channel) {
	c.mu.Lock()
	c.channel = ch
	c.appMessages = make(chan appMsg, 16)
	c.mu.Unlock()

	// Start the background message pump, LAN listener, and discovery.
	go c.encryptedPump(ch)
	c.startLANListener()
	go c.offerLANAddresses()
}

// Send writes a message to the peer via the preferred transport.
// In raw mode, data is sent as-is. In encrypted mode, data is treated
// as plaintext — it is framed and encrypted automatically.
func (c *Conn) Send(ctx context.Context, mt websocket.MessageType, data []byte) error {
	c.mu.Lock()
	ch := c.channel
	t := c.transports[c.preferred]
	c.mu.Unlock()

	if ch != nil {
		framed := make([]byte, 1+len(data))
		framed[0] = msgApp
		copy(framed[1:], data)
		encrypted := ch.Encrypt(framed)
		return t.ws.Write(ctx, websocket.MessageBinary, encrypted)
	}
	return t.ws.Write(ctx, mt, data)
}

// Recv reads the next application message from the peer. In raw mode,
// returns raw bytes from the transport. In encrypted mode, the background
// pump handles decryption, reordering, and control messages — Recv just
// delivers application messages.
func (c *Conn) Recv(ctx context.Context) (websocket.MessageType, []byte, error) {
	c.mu.Lock()
	ch := c.channel
	apps := c.appMessages
	c.mu.Unlock()

	if ch != nil {
		// Encrypted mode — read from the pump's delivery channel.
		select {
		case msg := <-apps:
			return msg.mt, msg.data, msg.err
		case <-ctx.Done():
			return 0, nil, ctx.Err()
		case <-c.ctx.Done():
			return 0, nil, c.ctx.Err()
		}
	}

	// Raw mode — read directly from transports.
	for {
		select {
		case msg := <-c.incoming:
			if msg.err != nil {
				if c.hasActiveTransports(msg.transportIdx) {
					slog.Info("transport closed", "name", c.transports[msg.transportIdx].name)
					continue
				}
				return 0, nil, msg.err
			}
			return msg.mt, msg.data, nil
		case <-ctx.Done():
			return 0, nil, ctx.Err()
		case <-c.ctx.Done():
			return 0, nil, c.ctx.Err()
		}
	}
}

// encryptedPump is the background message processor for encrypted mode.
// It reads from all transports, handles decryption, reordering, and
// control messages, and delivers application messages to appMessages.
func (c *Conn) encryptedPump(ch *crypto.Channel) {
	for {
		// Check the reorder buffer first.
		if mt, data, ok := c.tryDeliverBuffered(ch); ok {
			select {
			case c.appMessages <- appMsg{mt: mt, data: data}:
			case <-c.ctx.Done():
				return
			}
			continue
		}

		// Read from any transport.
		select {
		case msg := <-c.incoming:
			if msg.err != nil {
				if c.hasActiveTransports(msg.transportIdx) {
					slog.Info("transport closed", "name", c.transports[msg.transportIdx].name)
					continue
				}
				// All transports failed — signal the error.
				select {
				case c.appMessages <- appMsg{err: msg.err}:
				case <-c.ctx.Done():
				}
				return
			}

			mt, data, delivered := c.processEncrypted(ch, msg)
			if delivered {
				select {
				case c.appMessages <- appMsg{mt: mt, data: data}:
				case <-c.ctx.Done():
					return
				}
			}

		case <-c.ctx.Done():
			return
		}
	}
}

// processEncrypted handles an incoming encrypted message. Returns the
// decrypted application payload if it's the next in sequence and is an
// application message. Returns delivered=false if the message was
// buffered (out of order) or was an internal control message.
func (c *Conn) processEncrypted(ch *crypto.Channel, msg incomingMsg) (websocket.MessageType, []byte, bool) {
	if len(msg.data) < 8 {
		return 0, nil, false // too short, ignore
	}

	seq := binary.LittleEndian.Uint64(msg.data[:8])

	c.mu.Lock()
	defer c.mu.Unlock()

	if seq != c.nextRecvSeq {
		// Out of order — buffer for later.
		c.reorderBuf[seq] = reorderEntry{mt: msg.mt, data: msg.data}
		return 0, nil, false
	}

	// In order — decrypt and process.
	return c.decryptAndDeliver(ch, msg.mt, msg.data)
}

// tryDeliverBuffered checks if the reorder buffer has the next expected
// message and delivers it. Must be called without holding c.mu.
func (c *Conn) tryDeliverBuffered(ch *crypto.Channel) (websocket.MessageType, []byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.reorderBuf[c.nextRecvSeq]
	if !ok {
		return 0, nil, false
	}
	delete(c.reorderBuf, c.nextRecvSeq)
	return c.decryptAndDeliver(ch, entry.mt, entry.data)
}

// decryptAndDeliver decrypts a message and handles it. For application
// messages, returns the plaintext. For control messages, processes them
// internally and returns delivered=false. Caller must hold c.mu.
func (c *Conn) decryptAndDeliver(ch *crypto.Channel, mt websocket.MessageType, data []byte) (websocket.MessageType, []byte, bool) {
	plaintext, err := ch.Decrypt(data)
	if err != nil {
		slog.Warn("decrypt failed", "err", err)
		c.nextRecvSeq++
		return 0, nil, false
	}
	c.nextRecvSeq++

	if len(plaintext) == 0 {
		return 0, nil, false
	}

	msgType := plaintext[0]
	payload := plaintext[1:]

	switch msgType {
	case msgApp:
		return mt, payload, true
	case msgLANOffer:
		c.handleLANOfferLocked(payload)
		return 0, nil, false
	case msgCutover:
		c.handleCutoverLocked(mt)
		return 0, nil, false
	default:
		slog.Warn("unknown control message type", "type", msgType)
		return 0, nil, false
	}
}

// handleLANOfferLocked processes a LAN address offer from the peer.
// Attempts a direct WebSocket connection in the background.
// Caller must hold c.mu.
func (c *Conn) handleLANOfferLocked(payload []byte) {
	if c.lanAttempted {
		return // already tried
	}
	c.lanAttempted = true

	var offer struct {
		Addrs []string `json:"addrs"`
		Port  string   `json:"port"`
	}
	if err := json.Unmarshal(payload, &offer); err != nil {
		slog.Warn("invalid LAN offer", "err", err)
		return
	}

	go c.attemptLANConnect(offer.Addrs, offer.Port)
}

// handleCutoverLocked processes a CUTOVER control message from the peer.
// If we have a LAN transport, we switch to it. Caller must hold c.mu.
func (c *Conn) handleCutoverLocked(mt websocket.MessageType) {
	// The peer sent CUTOVER, meaning they've switched to LAN. If we
	// have a LAN transport, switch to it too.
	slog.Info("received cutover from peer")

	for i, t := range c.transports {
		if t.name == "lan" && i != c.preferred {
			c.preferred = i
			slog.Info("switched preferred transport to LAN (peer cutover)")
			return
		}
	}
	slog.Info("received cutover but no LAN transport available yet")
}

// offerLANAddresses sends our LAN addresses to the peer as a control
// message through the encrypted channel.
func (c *Conn) offerLANAddresses() {
	ip := qr.LanIP()
	if ip == "localhost" {
		return // no LAN address available
	}

	c.mu.Lock()
	port := c.lanPort
	ch := c.channel
	t := c.transports[c.preferred]
	c.mu.Unlock()

	if ch == nil || port == 0 {
		return // no channel or no listener
	}

	offer, _ := json.Marshal(struct {
		Addrs []string `json:"addrs"`
		Port  string   `json:"port"`
	}{
		Addrs: []string{ip},
		Port:  fmt.Sprintf("%d", port),
	})

	framed := make([]byte, 1+len(offer))
	framed[0] = msgLANOffer
	copy(framed[1:], offer)

	encrypted := ch.Encrypt(framed)
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()
	if err := t.ws.Write(ctx, websocket.MessageBinary, encrypted); err != nil {
		slog.Warn("failed to send LAN offer", "err", err)
	}
}

// attemptLANConnect tries to establish a direct WebSocket connection
// to the peer's LAN addresses. If successful, adds it as a transport
// and initiates cutover from relay to LAN.
func (c *Conn) attemptLANConnect(addrs []string, port string) {
	ctx, cancel := context.WithTimeout(c.ctx, 3*time.Second)
	defer cancel()

	for _, addr := range addrs {
		// Check if we already have a LAN transport (accept side won the race).
		if c.hasLANTransport() {
			return
		}

		url := fmt.Sprintf("ws://%s:%s/direct", addr, port)
		slog.Info("attempting LAN connection", "url", url)

		ws, _, err := websocket.Dial(ctx, url, nil)
		if err != nil {
			slog.Info("LAN dial failed", "addr", addr, "err", err)
			continue
		}

		// Check again after dial — the accept side may have added one
		// while we were connecting.
		if c.hasLANTransport() {
			ws.CloseNow()
			return
		}

		// Connection succeeded. Add as transport and cut over.
		idx := c.addTransport(ws, "lan")
		slog.Info("LAN transport added (dialed)", "idx", idx, "addr", addr)

		c.sendCutover(idx)
		return
	}

	slog.Info("LAN connection failed for all addresses")
}

// hasLANTransport returns true if a LAN transport has already been added.
func (c *Conn) hasLANTransport() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, t := range c.transports {
		if t.name == "lan" {
			return true
		}
	}
	return false
}

// startLANListener starts a local WebSocket server on an ephemeral port.
// When a peer connects to /direct, it is added as a new transport.
func (c *Conn) startLANListener() {
	ln, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		slog.Warn("failed to start LAN listener", "err", err)
		return
	}

	c.mu.Lock()
	c.lanListener = ln
	c.lanPort = ln.Addr().(*net.TCPAddr).Port
	c.mu.Unlock()

	slog.Info("LAN listener started", "port", c.lanPort)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /direct", func(w http.ResponseWriter, r *http.Request) {
		// Reject if we already have a LAN transport.
		if c.hasLANTransport() {
			http.Error(w, "already connected", http.StatusConflict)
			return
		}

		ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			slog.Warn("LAN accept failed", "err", err)
			return
		}

		// The peer connected directly. Add as a transport.
		// Authentication happens implicitly: subsequent messages
		// on this transport are encrypted, and decrypt will fail
		// if the peer doesn't share our channel keys.
		idx := c.addTransport(ws, "lan")
		slog.Info("LAN peer accepted", "idx", idx)

		// Don't cut over yet — wait for the connecting side to
		// send cutover first. When we receive their cutover
		// message (through normal message processing), we know
		// they've committed to the LAN transport. Then we also
		// switch via handleCutoverLocked -> switchToLAN.
	})

	srv := &http.Server{Handler: mux}
	go func() {
		if err := srv.Serve(ln); err != nil && c.ctx.Err() == nil {
			slog.Info("LAN listener closed", "err", err)
		}
	}()

	// Shut down when the conn closes.
	go func() {
		<-c.ctx.Done()
		srv.Close()
	}()
}

// sendCutover sends a CUTOVER control message on the current preferred
// transport, then switches to the new transport for subsequent sends.
func (c *Conn) sendCutover(newPreferred int) {
	c.mu.Lock()
	ch := c.channel
	t := c.transports[c.preferred]
	c.mu.Unlock()

	if ch == nil {
		return
	}

	framed := []byte{msgCutover}
	encrypted := ch.Encrypt(framed)

	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()
	if err := t.ws.Write(ctx, websocket.MessageBinary, encrypted); err != nil {
		slog.Warn("failed to send cutover", "err", err)
		return
	}

	c.mu.Lock()
	c.preferred = newPreferred
	c.mu.Unlock()

	slog.Info("cutover sent, switched transport",
		"from", t.name,
		"to", c.transports[newPreferred].name)
}

// addTransport adds a new transport and starts its reader goroutine.
func (c *Conn) addTransport(ws *websocket.Conn, name string) int {
	c.mu.Lock()
	idx := len(c.transports)
	c.transports = append(c.transports, &transport{ws: ws, name: name})
	c.mu.Unlock()

	go c.readLoop(idx, ws)
	return idx
}

// hasActiveTransports returns true if there are other transports
// besides the one at the given index that might still deliver messages.
func (c *Conn) hasActiveTransports(failedIdx int) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, t := range c.transports {
		if i != failedIdx && !t.cutover {
			return true
		}
	}
	return false
}

// Close gracefully closes all transports and the LAN listener.
func (c *Conn) Close() error {
	c.cancel()
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lanListener != nil {
		c.lanListener.Close()
	}
	var firstErr error
	for _, t := range c.transports {
		if err := t.ws.Close(websocket.StatusNormalClosure, ""); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// CloseNow immediately closes all transports and the LAN listener.
func (c *Conn) CloseNow() error {
	c.cancel()
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lanListener != nil {
		c.lanListener.Close()
	}
	var firstErr error
	for _, t := range c.transports {
		if err := t.ws.CloseNow(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// PreferredTransport returns the name of the currently preferred transport.
func (c *Conn) PreferredTransport() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.transports[c.preferred].name
}

// SetReadLimit sets the maximum message size on all transports.
func (c *Conn) SetReadLimit(n int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, t := range c.transports {
		t.ws.SetReadLimit(n)
	}
}
