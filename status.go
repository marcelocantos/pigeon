// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"encoding/json"
	"net/http"
	"time"
)

// Version is the relay build version string. Overridden at link time via
// -ldflags "-X github.com/marcelocantos/pigeon.Version=<tag>" or assigned
// from main.version after flag parse. Defaults to "dev" for local builds.
var Version = "dev"

// Commit is an optional source revision (typically a git SHA). Overridden
// at link time via -ldflags "-X github.com/marcelocantos/pigeon.Commit=<sha>".
// Empty when not injected.
var Commit = ""

var startedAt = time.Now().UTC()

// Status is the JSON body of GET /status — a general process diagnostic,
// not a probe. Keep GET /health minimal for load balancers and WakeRelay.
type Status struct {
	Service   string `json:"service"`
	Status    string `json:"status"`
	Version   string `json:"version"`
	Commit    string `json:"commit,omitempty"`
	StartedAt string `json:"started_at"`
	UptimeSec int64  `json:"uptime_sec"`
}

// CurrentStatus returns a point-in-time process status snapshot.
func CurrentStatus() Status {
	now := time.Now().UTC()
	return Status{
		Service:   "pigeon",
		Status:    "ok",
		Version:   Version,
		Commit:    Commit,
		StartedAt: startedAt.Format(time.RFC3339),
		UptimeSec: int64(now.Sub(startedAt).Seconds()),
	}
}

// HandleHealth serves GET /health — liveness only.
func HandleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// HandleStatus serves GET /status — version, uptime, and related process info.
func HandleStatus(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(CurrentStatus())
}
