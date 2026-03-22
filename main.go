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
	"context"
	"crypto/rand"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/coder/websocket"
)

type instance struct {
	id   string
	conn *websocket.Conn
	ctx  context.Context
	mu   sync.Mutex
}

type relay struct {
	mu        sync.RWMutex
	instances map[string]*instance
}

func newRelay() *relay {
	return &relay{instances: make(map[string]*instance)}
}

func (r *relay) register(inst *instance) {
	r.mu.Lock()
	r.instances[inst.id] = inst
	r.mu.Unlock()
}

func (r *relay) unregister(id string) {
	r.mu.Lock()
	delete(r.instances, id)
	r.mu.Unlock()
}

func (r *relay) get(id string) *instance {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.instances[id]
}

func generateID() string {
	// 4 bytes → base36 ≈ 6-7 chars. Plenty for session uniqueness
	// (~4 billion possibilities) without being unwieldy in URLs.
	b := make([]byte, 4)
	rand.Read(b)
	n := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	return strconv.FormatUint(uint64(n), 36)
}

// registerRoutes sets up HTTP and WebSocket handlers on mux.
func registerRoutes(mux *http.ServeMux, r *relay) {
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Backend registers here. The relay assigns an instance ID and sends
	// it back as the first text message. Then the connection stays open
	// for bidirectional bridging with clients.
	mux.HandleFunc("GET /register", func(w http.ResponseWriter, req *http.Request) {
		conn, err := websocket.Accept(w, req, &websocket.AcceptOptions{
			OriginPatterns: []string{"*"},
		})
		if err != nil {
			slog.Error("register: accept failed", "err", err)
			return
		}
		defer conn.CloseNow()

		ctx := req.Context()
		id := generateID()

		// Send the instance ID back to the backend.
		if err := conn.Write(ctx, websocket.MessageText, []byte(id)); err != nil {
			return
		}

		inst := &instance{id: id, conn: conn, ctx: ctx}
		r.register(inst)
		defer r.unregister(id)

		slog.Info("instance registered", "id", id)

		// Keep alive until backend disconnects. Backend→client forwarding
		// is handled by the client bridge goroutine reading from this conn.
		<-ctx.Done()
		slog.Info("instance disconnected", "id", id)
	})

	// Client connects here to reach a specific backend instance.
	mux.HandleFunc("GET /ws/{id}", func(w http.ResponseWriter, req *http.Request) {
		instanceID := req.PathValue("id")
		inst := r.get(instanceID)
		if inst == nil {
			http.Error(w, `{"error":"instance not found"}`, http.StatusNotFound)
			return
		}

		clientConn, err := websocket.Accept(w, req, &websocket.AcceptOptions{
			OriginPatterns: []string{"*"},
		})
		if err != nil {
			slog.Error("client: accept failed", "err", err)
			return
		}
		defer clientConn.CloseNow()

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
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	r := newRelay()
	mux := http.NewServeMux()
	registerRoutes(mux, r)

	addr := ":" + port
	slog.Info("tern starting", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("tern failed", "err", err)
		os.Exit(1)
	}
}
