// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package httpguard

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const relayPort = "443"

// TestRejectsRebindingHost is the DNS-rebinding oracle: a request whose
// Host names a page the attacker controls must not be treated as
// addressed to this relay.
func TestRejectsRebindingHost(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/pigeon", nil)
	req.Host = "evil.example:" + relayPort
	if AllowRequest(req, relayPort, "") {
		t.Fatal("Host: evil.example allowed, want reject")
	}
}

func TestAcceptsLoopbackHost(t *testing.T) {
	for _, host := range []string{
		"127.0.0.1:" + relayPort,
		"[::1]:" + relayPort,
		"localhost:" + relayPort,
		"LocalHost:" + relayPort,
	} {
		req := httptest.NewRequest(http.MethodGet, "/pigeon", nil)
		req.Host = host
		if !AllowRequest(req, relayPort, "") {
			t.Errorf("Host: %s rejected, want accept", host)
		}
	}
}

func TestAcceptsConfiguredDomain(t *testing.T) {
	const domain = "carrier-pigeon.fly.dev"
	req := httptest.NewRequest(http.MethodGet, "/pigeon", nil)
	req.Host = domain
	if !AllowRequest(req, relayPort, domain) {
		t.Fatalf("Host: %s rejected with domain configured", domain)
	}
	req.Host = "evil.example"
	if AllowRequest(req, relayPort, domain) {
		t.Fatal("Host: evil.example allowed when domain is configured")
	}
}

func TestRejectsCrossOriginNullAndFile(t *testing.T) {
	for _, origin := range []string{"null", "file://", "data:text/html,x"} {
		req := httptest.NewRequest(http.MethodGet, "/pigeon", nil)
		req.Host = "127.0.0.1:" + relayPort
		req.Header.Set("Origin", origin)
		if AllowRequest(req, relayPort, "") {
			t.Errorf("Origin: %s allowed, want reject", origin)
		}
	}
}

func TestAcceptsBrowserOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/pigeon", nil)
	req.Host = "127.0.0.1:" + relayPort
	req.Header.Set("Origin", "https://app.example")
	if !AllowRequest(req, relayPort, "") {
		t.Fatal("https Origin from a client app rejected; the relay is a public rendezvous")
	}
}

func TestAbsentOriginIsAccepted(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/pigeon", nil)
	req.Host = "127.0.0.1:" + relayPort
	if !AllowRequest(req, relayPort, "") {
		t.Fatal("missing Origin rejected; native clients omit it")
	}
}

func TestWriteDiagCORSReflectsOriginNotStar(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Host = "127.0.0.1:" + relayPort
	req.Header.Set("Origin", "https://app.example")
	rec := httptest.NewRecorder()
	WriteDiagCORS(rec, req)
	got := rec.Header().Get("Access-Control-Allow-Origin")
	if got == "*" {
		t.Fatal("CORS wrote *; want the request Origin reflected")
	}
	if got != "https://app.example" {
		t.Fatalf("CORS Origin = %q, want the request Origin", got)
	}
}

func TestWriteDiagCORSOmitsNullOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Host = "127.0.0.1:" + relayPort
	req.Header.Set("Origin", "null")
	rec := httptest.NewRecorder()
	WriteDiagCORS(rec, req)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("CORS for Origin: null = %q, want empty", got)
	}
}
