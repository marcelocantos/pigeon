#!/usr/bin/env bash
# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0
#
# Measures combined unit + integration test coverage.
#
# Phase A: Normal unit tests (server code runs in-process, captures
#          client + library coverage).
# Phase B: Tests against a coverage-instrumented relay binary
#          (captures cmd/pigeon server code coverage).
# Final:   Merge both profiles for the combined picture.
#
# Usage:
#   ./scripts/integration-coverage.sh

set -uo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

COVERDIR=$(mktemp -d)
INTEGRATION_COVERDIR="$COVERDIR/integration"
mkdir -p "$INTEGRATION_COVERDIR"

RELAY_PID=""

cleanup() {
    if [[ -n "$RELAY_PID" ]]; then
        kill "$RELAY_PID" 2>/dev/null || true
        wait "$RELAY_PID" 2>/dev/null || true
    fi
    rm -rf "$COVERDIR"
}
trap cleanup EXIT

# --- Phase A: Unit tests (in-process server) ---
echo "=== Phase A: Unit tests (in-process relay) ==="
go test -count=1 -timeout=180s -coverprofile="$COVERDIR/unit.out" ./... 2>&1 \
    | grep -E "^ok |^FAIL|coverage" || true
echo

# --- Phase B: Tests against instrumented relay binary ---
echo "=== Phase B: Build coverage-instrumented relay ==="
go build -cover -o "$COVERDIR/pigeon-cover" ./cmd/pigeon || exit 1

QUIC_PORT=$(python3 -c "import socket; s=socket.socket(socket.AF_INET,socket.SOCK_DGRAM); s.bind(('',0)); print(s.getsockname()[1]); s.close()")
WT_PORT=$(python3 -c "import socket; s=socket.socket(socket.AF_INET,socket.SOCK_DGRAM); s.bind(('',0)); print(s.getsockname()[1]); s.close()")

echo "Starting instrumented relay (QUIC=$QUIC_PORT, WT=$WT_PORT)..."
GOCOVERDIR="$INTEGRATION_COVERDIR" "$COVERDIR/pigeon-cover" \
    --port "$WT_PORT" --quic-port "$QUIC_PORT" 2>/dev/null &
RELAY_PID=$!
sleep 2

if ! kill -0 "$RELAY_PID" 2>/dev/null; then
    echo "ERROR: relay failed to start"
    exit 1
fi

echo "Running tests against instrumented relay..."
PIGEON_TEST_QUIC_URL="https://127.0.0.1:$QUIC_PORT" \
PIGEON_TEST_WT_URL="https://127.0.0.1:$WT_PORT" \
go test -count=1 -timeout=180s . 2>&1 \
    | grep -E "^ok |^FAIL" || true
echo

echo "Stopping relay..."
kill "$RELAY_PID" 2>/dev/null || true
wait "$RELAY_PID" 2>/dev/null || true
RELAY_PID=""
echo

# --- Report ---
echo "============================================"
echo "             COVERAGE REPORT"
echo "============================================"
echo

echo "--- Unit tests (in-process) ---"
go tool cover -func="$COVERDIR/unit.out" | tail -1
echo

if ls "$INTEGRATION_COVERDIR"/* 1>/dev/null 2>&1; then
    echo "--- Server binary (instrumented) ---"
    go tool covdata percent -i="$INTEGRATION_COVERDIR"
    go tool covdata textfmt -i="$INTEGRATION_COVERDIR" -o "$COVERDIR/server.out"
    echo

    echo "--- Merged total ---"
    head -1 "$COVERDIR/unit.out" > "$COVERDIR/merged.out"
    tail -n+2 "$COVERDIR/unit.out" >> "$COVERDIR/merged.out"
    tail -n+2 "$COVERDIR/server.out" >> "$COVERDIR/merged.out"
    go tool cover -func="$COVERDIR/merged.out" | tail -1
    echo

    go tool cover -html="$COVERDIR/merged.out" -o /tmp/pigeon-coverage.html
    echo "HTML report: open /tmp/pigeon-coverage.html"
else
    echo "(No server-side coverage data collected)"
fi
