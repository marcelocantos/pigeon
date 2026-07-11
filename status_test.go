// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCurrentStatusShape(t *testing.T) {
	prevV, prevC := Version, Commit
	t.Cleanup(func() { Version, Commit = prevV, prevC })
	Version = "v0.30.0"
	Commit = "abc1234"

	st := CurrentStatus()
	if st.Service != "pigeon" {
		t.Fatalf("service: %q", st.Service)
	}
	if st.Status != "ok" {
		t.Fatalf("status: %q", st.Status)
	}
	if st.Version != "v0.30.0" {
		t.Fatalf("version: %q", st.Version)
	}
	if st.Commit != "abc1234" {
		t.Fatalf("commit: %q", st.Commit)
	}
	if _, err := time.Parse(time.RFC3339, st.StartedAt); err != nil {
		t.Fatalf("started_at: %v", err)
	}
	if st.UptimeSec < 0 {
		t.Fatalf("uptime_sec: %d", st.UptimeSec)
	}
}

func TestHandleHealth(t *testing.T) {
	rr := httptest.NewRecorder()
	HandleHealth(rr, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("code %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type %q", ct)
	}
	if body := rr.Body.String(); body != `{"status":"ok"}` {
		t.Fatalf("body %q", body)
	}
}

func TestHandleStatus(t *testing.T) {
	prevV, prevC := Version, Commit
	t.Cleanup(func() { Version, Commit = prevV, prevC })
	Version = "v9.9.9"
	Commit = "deadbeef"

	rr := httptest.NewRecorder()
	HandleStatus(rr, httptest.NewRequest(http.MethodGet, "/status", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("code %d", rr.Code)
	}
	var st Status
	if err := json.Unmarshal(rr.Body.Bytes(), &st); err != nil {
		t.Fatal(err)
	}
	if st.Version != "v9.9.9" || st.Commit != "deadbeef" || st.Service != "pigeon" {
		t.Fatalf("decoded %+v", st)
	}
}
