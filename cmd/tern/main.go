// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Command tern is a minimal WebSocket relay server. Backend instances
// register and receive a unique instance ID. Clients connect by ID
// and all traffic is forwarded bidirectionally.
//
// Endpoints:
//
//	GET /health             — health check
//	GET /register           — backend connects here (WebSocket upgrade)
//	GET /ws/<instance-id>   — client connects here (WebSocket upgrade)
package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/coder/websocket"
	"github.com/marcelocantos/tern"
)

// version is set at build time via -ldflags "-X main.version=v0.1.0".
var version = "dev"

// maxMessageSize is the WebSocket read limit for both backend and client
// connections. 1 MiB is generous for relay traffic (typically small JSON
// messages or encrypted binary frames) while preventing a single frame
// from consuming unbounded memory.
const maxMessageSize = 1 << 20 // 1 MiB

type instance struct {
	id       string
	conn     *websocket.Conn
	ctx      context.Context
	mu       sync.Mutex
	occupied bool // true when a client is connected; protected by mu
}

type hub struct {
	mu        sync.RWMutex
	instances map[string]*instance
}

func newHub() *hub {
	return &hub{instances: make(map[string]*instance)}
}

func (h *hub) register(inst *instance) {
	h.mu.Lock()
	h.instances[inst.id] = inst
	h.mu.Unlock()
}

func (h *hub) unregister(id string) {
	h.mu.Lock()
	delete(h.instances, id)
	h.mu.Unlock()
}

func (h *hub) get(id string) *instance {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.instances[id]
}

func generateID() string {
	// 4 bytes → base36 ≈ 6-7 chars. Plenty for session uniqueness
	// (~4 billion possibilities) without being unwieldy in URLs.
	// NOTE: 32-bit IDs are adequate for moderate-traffic relays where
	// instances are short-lived and the chance of collision is low. For
	// high-traffic public deployments, increase to 8 bytes (64-bit) to
	// reduce collision probability.
	b := make([]byte, 4)
	rand.Read(b)
	n := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	return strconv.FormatUint(uint64(n), 36)
}

// registerRoutes sets up HTTP and WebSocket handlers on mux.
// If token is non-empty, /register requires a matching Bearer token.
func registerRoutes(mux *http.ServeMux, h *hub, token string) {
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Backend registers here. The relay assigns an instance ID and sends
	// it back as the first text message. Then the connection stays open
	// for bidirectional bridging with clients.
	mux.HandleFunc("GET /register", func(w http.ResponseWriter, req *http.Request) {
		if token != "" {
			auth := req.Header.Get("Authorization")
			if auth != "Bearer "+token {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
		}
		conn, err := websocket.Accept(w, req, &websocket.AcceptOptions{
			// CORS wildcard is intentional: the relay bridges arbitrary
			// backends and clients that may run on any origin. Deployers
			// who want tighter control should restrict this list.
			OriginPatterns: []string{"*"},
		})
		if err != nil {
			slog.Error("register: accept failed", "err", err)
			return
		}
		defer conn.CloseNow()
		conn.SetReadLimit(maxMessageSize)

		ctx := req.Context()
		id := generateID()

		// Send the instance ID back to the backend.
		if err := conn.Write(ctx, websocket.MessageText, []byte(id)); err != nil {
			return
		}

		inst := &instance{id: id, conn: conn, ctx: ctx}
		h.register(inst)
		defer h.unregister(id)

		slog.Info("instance registered", "id", id)

		// Keep alive until backend disconnects. Backend→client forwarding
		// is handled by the client bridge goroutine reading from this conn.
		<-ctx.Done()
		slog.Info("instance disconnected", "id", id)
	})

	// Client connects here to reach a specific backend instance.
	mux.HandleFunc("GET /ws/{id}", func(w http.ResponseWriter, req *http.Request) {
		instanceID := req.PathValue("id")
		inst := h.get(instanceID)
		if inst == nil {
			http.Error(w, `{"error":"instance not found"}`, http.StatusNotFound)
			return
		}

		// Enforce single client per instance to prevent concurrent reads
		// on the backend connection (data race).
		inst.mu.Lock()
		if inst.occupied {
			inst.mu.Unlock()
			http.Error(w, `{"error":"instance already has a connected client"}`, http.StatusConflict)
			return
		}
		inst.occupied = true
		inst.mu.Unlock()
		defer func() {
			inst.mu.Lock()
			inst.occupied = false
			inst.mu.Unlock()
		}()

		clientConn, err := websocket.Accept(w, req, &websocket.AcceptOptions{
			// CORS wildcard is intentional: the relay bridges arbitrary
			// backends and clients that may run on any origin. Deployers
			// who want tighter control should restrict this list.
			OriginPatterns: []string{"*"},
		})
		if err != nil {
			slog.Error("client: accept failed", "err", err)
			return
		}
		defer clientConn.CloseNow()
		clientConn.SetReadLimit(maxMessageSize)

		ctx := req.Context()
		slog.Info("client connected", "instance", instanceID)

		// Bridge: bidirectional forwarding between client and backend.

		// backend → client
		go func() {
			for {
				mt, data, err := inst.conn.Read(inst.ctx)
				if err != nil {
					clientConn.Close(websocket.StatusGoingAway, "instance disconnected")
					return
				}
				if err := clientConn.Write(ctx, mt, data); err != nil {
					return
				}
			}
		}()

		// client → backend
		for {
			mt, data, err := clientConn.Read(ctx)
			if err != nil {
				slog.Info("client disconnected", "instance", instanceID)
				return
			}
			inst.mu.Lock()
			err = inst.conn.Write(inst.ctx, mt, data)
			inst.mu.Unlock()
			if err != nil {
				slog.Warn("forward to instance failed", "instance", instanceID, "err", err)
				return
			}
		}
	})
}

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	helpAgent := flag.Bool("help-agent", false, "print help and agent guide")
	port := flag.String("port", "", "listening port (overrides PORT env var)")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if *helpAgent {
		var buf bytes.Buffer
		flag.CommandLine.SetOutput(&buf)
		flag.Usage()
		fmt.Print(buf.String())
		fmt.Println(tern.AgentGuide)
		os.Exit(0)
	}

	listenPort := *port
	if listenPort == "" {
		listenPort = os.Getenv("PORT")
	}
	if listenPort == "" {
		listenPort = "8080"
	}

	// TERN_TOKEN restricts /register to authorized backends.
	// If unset, registration is open.
	token := os.Getenv("TERN_TOKEN")

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	h := newHub()
	mux := http.NewServeMux()
	registerRoutes(mux, h, token)

	addr := ":" + listenPort
	slog.Info("tern starting", "addr", addr, "version", version)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("tern failed", "err", err)
		os.Exit(1)
	}
}
