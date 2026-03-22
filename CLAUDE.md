# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Tern is a minimal WebSocket relay server (Go) that bridges backend instances and clients. Extracted from the jevon project. Deployed on Fly.io (Sydney region).

## Build & Run

```bash
go build -o tern .                # build
./tern                            # run (listens on :8080)
PORT=3000 ./tern                  # custom port
docker build -t tern . && docker run -p 8080:8080 tern  # Docker
```

No test suite or Makefile exists yet.

## Architecture

Single-file app (`main.go`, ~180 lines) with one dependency (`github.com/coder/websocket`).

**Core types:**
- `relay` — thread-safe registry of active backend instances (RWMutex-protected map)
- `instance` — a registered backend: ID, WebSocket conn, context, write mutex

**Endpoints:**
- `GET /health` — JSON health check
- `GET /register` — backend connects via WebSocket, receives a random base36 ID, stays open
- `GET /ws/{id}` — client connects via WebSocket, bidirectional bridge to the backend instance

**Flow:** Backend registers → gets ID → client connects with that ID → two goroutines bridge traffic in both directions. Per-instance mutex serialises writes to the backend connection.

## Deployment

Fly.io config in `fly.toml`. Multi-stage Dockerfile (`golang:1.25-alpine` → `alpine:3.21`). Shared-cpu-1x, 256MB, auto-start/stop with zero minimum machines.

## Delivery

Merged to master.
