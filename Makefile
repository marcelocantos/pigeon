# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0

JDK21 ?= /opt/homebrew/Cellar/openjdk@21/21.0.10/libexec/openjdk.jdk/Contents/Home

.PHONY: all build test test-go test-swift test-kotlin test-web \
        test-live bench clean

# --- Build ---

all: build

build: build-go build-swift

build-go:
	go build ./...

build-swift:
	swift build

# --- Tests (local only, no relay token needed) ---

test: test-go test-swift test-kotlin test-web

test-go:
	go test -count=1 -timeout=60s ./...

test-swift:
	swift test

test-kotlin:
	JAVA_HOME=$(JDK21) android/gradlew \
		-p $(CURDIR)/android test --no-daemon --console=plain

test-web:
	cd web && npx tsx --test src/crypto.test.ts

# --- Live tests (require TERN_TOKEN) ---

test-live: test-go-live

test-go-live:
ifndef TERN_TOKEN
	$(error TERN_TOKEN is required for live tests)
endif
	TERN_TOKEN=$(TERN_TOKEN) go test -count=1 -timeout=120s -v ./... 2>&1 \
		| grep -E '^\s*(=== RUN|--- |ok |FAIL)'

# --- Benchmarks ---

bench:
	go test -bench=. -benchtime=2s -count=1 -timeout=120s -run=^$$ .

bench-live:
ifndef TERN_TOKEN
	$(error TERN_TOKEN is required for live benchmarks)
endif
	TERN_TOKEN=$(TERN_TOKEN) go test -bench=. -benchtime=2s -count=1 -timeout=120s -run=^$$ .

# --- Server ---

server:
	go run ./cmd/tern

# --- Code generation ---

generate:
	go run ./cmd/protogen protocol/pairing.yaml

# --- Clean ---

clean:
	rm -rf .build/
	rm -f tern tern-test-binary
	go clean -testcache
