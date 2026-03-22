// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func startTestRelay(t *testing.T) *httptest.Server {
	t.Helper()
	r := newRelay()
	mux := http.NewServeMux()
	registerRoutes(mux, r)
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return ts
}

func wsURL(ts *httptest.Server, path string) string {
	return "ws" + strings.TrimPrefix(ts.URL, "http") + path
}

func dial(t *testing.T, url string) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatalf("WebSocket dial %s failed: %v", url, err)
	}
	return conn
}

// registerBackend connects to /register, reads the assigned instance ID, and
// returns the connection and ID. The connection is closed on test cleanup.
func registerBackend(t *testing.T, ts *httptest.Server) (*websocket.Conn, string) {
	t.Helper()
	conn := dial(t, wsURL(ts, "/register"))
	t.Cleanup(func() { conn.CloseNow() })

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	mt, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("Failed to read instance ID: %v", err)
	}
	if mt != websocket.MessageText {
		t.Fatalf("Expected text message for instance ID, got %v", mt)
	}
	id := string(data)
	if id == "" {
		t.Fatal("Received empty instance ID")
	}
	return conn, id
}

func TestHealthEndpoint(t *testing.T) {
	ts := startTestRelay(t)

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status = %v, want ok", body["status"])
	}
}

func TestRegisterAssignsID(t *testing.T) {
	ts := startTestRelay(t)
	_, id := registerBackend(t, ts)
	t.Logf("Assigned instance ID: %s", id)
}

func TestClientConnectUnknownInstance(t *testing.T) {
	ts := startTestRelay(t)

	resp, err := http.Get(ts.URL + "/ws/nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestBidirectionalBridge(t *testing.T) {
	ts := startTestRelay(t)

	backend, id := registerBackend(t, ts)

	// Client connects to the backend instance.
	client := dial(t, wsURL(ts, "/ws/"+id))
	defer client.CloseNow()

	ctx := context.Background()

	// client → backend
	if err := client.Write(ctx, websocket.MessageText, []byte("hello from client")); err != nil {
		t.Fatalf("Client write failed: %v", err)
	}

	rctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	mt, data, err := backend.Read(rctx)
	if err != nil {
		t.Fatalf("Backend read failed: %v", err)
	}
	if mt != websocket.MessageText || string(data) != "hello from client" {
		t.Fatalf("Backend got (%v, %q), want (text, %q)", mt, data, "hello from client")
	}

	// backend → client
	if err := backend.Write(ctx, websocket.MessageText, []byte("hello from backend")); err != nil {
		t.Fatalf("Backend write failed: %v", err)
	}

	rctx2, cancel2 := context.WithTimeout(ctx, 2*time.Second)
	defer cancel2()
	mt, data, err = client.Read(rctx2)
	if err != nil {
		t.Fatalf("Client read failed: %v", err)
	}
	if mt != websocket.MessageText || string(data) != "hello from backend" {
		t.Fatalf("Client got (%v, %q), want (text, %q)", mt, data, "hello from backend")
	}
}

func TestBinaryFrameForwarding(t *testing.T) {
	ts := startTestRelay(t)

	backend, id := registerBackend(t, ts)
	client := dial(t, wsURL(ts, "/ws/"+id))
	defer client.CloseNow()

	ctx := context.Background()
	payload := []byte{0xDE, 0xAD, 0xBE, 0xEF}

	// client → backend (binary)
	if err := client.Write(ctx, websocket.MessageBinary, payload); err != nil {
		t.Fatalf("Client binary write failed: %v", err)
	}

	rctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	mt, data, err := backend.Read(rctx)
	if err != nil {
		t.Fatalf("Backend read failed: %v", err)
	}
	if mt != websocket.MessageBinary {
		t.Fatalf("Expected binary message, got %v", mt)
	}
	if string(data) != string(payload) {
		t.Fatalf("Backend got %x, want %x", data, payload)
	}
}

func TestBackendDisconnectClosesClient(t *testing.T) {
	ts := startTestRelay(t)

	backend, id := registerBackend(t, ts)
	client := dial(t, wsURL(ts, "/ws/"+id))
	defer client.CloseNow()

	// Backend disconnects.
	backend.Close(websocket.StatusNormalClosure, "bye")

	// Client should get an error on next read.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, _, err := client.Read(ctx)
	if err == nil {
		t.Fatal("Expected client read to fail after backend disconnect")
	}
}

func TestMultipleInstances(t *testing.T) {
	ts := startTestRelay(t)

	_, id1 := registerBackend(t, ts)
	_, id2 := registerBackend(t, ts)

	if id1 == id2 {
		t.Fatalf("Two backends got the same ID: %s", id1)
	}
}

func TestSecondClientRejected(t *testing.T) {
	ts := startTestRelay(t)

	_, id := registerBackend(t, ts)

	// First client connects successfully.
	client1 := dial(t, wsURL(ts, "/ws/"+id))
	defer client1.CloseNow()

	// Second client should be rejected with 409 Conflict.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, resp, err := websocket.Dial(ctx, wsURL(ts, "/ws/"+id), nil)
	if err == nil {
		t.Fatal("second client dial should have failed")
	}
	if resp != nil && resp.StatusCode != http.StatusConflict {
		t.Fatalf("second client status = %d, want %d", resp.StatusCode, http.StatusConflict)
	}
}

func TestGenerateIDLength(t *testing.T) {
	for range 100 {
		id := generateID()
		if len(id) == 0 || len(id) > 7 {
			t.Errorf("generateID() = %q, want 1-7 chars", id)
		}
	}
}
