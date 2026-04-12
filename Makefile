# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0

JDK21 ?= /opt/homebrew/Cellar/openjdk@21/21.0.10/libexec/openjdk.jdk/Contents/Home

.PHONY: all build test test-go test-swift test-kotlin test-web \
        e2e e2e-go e2e-swift e2e-kotlin \
        test-live bench clean

# --- Build ---

all: build

build: build-go build-swift

build-go:
	go build ./...

build-swift:
	swift build

# --- Unit tests (local only, no relay needed) ---

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

# --- E2E tests (standalone, against local relay) ---

e2e: e2e-go e2e-swift e2e-kotlin

e2e-go:
	go test -count=1 -timeout=60s -run "TestStreamRoundTrip/local" .

e2e-swift:
	swift run pigeon-e2e-swift

e2e-kotlin:
	JAVA_HOME=$(JDK21) android/gradlew \
		-p $(CURDIR)/android :pigeon:test --no-daemon --console=plain \
		--tests "com.marcelocantos.pigeon.relay.PigeonConnE2ETest"

# --- E2E tests against live relay (require PIGEON_TOKEN) ---

e2e-live: e2e-go-live e2e-swift-live

e2e-go-live:
ifndef PIGEON_TOKEN
	$(error PIGEON_TOKEN is required for live E2E tests)
endif
	PIGEON_TOKEN=$(PIGEON_TOKEN) go test -count=1 -timeout=120s -v \
		-run "TestStreamRoundTrip/live" . 2>&1 \
		| grep -E '^\s*(=== RUN|--- |ok |FAIL)'

e2e-swift-live:
ifndef PIGEON_TOKEN
	$(error PIGEON_TOKEN is required for live E2E tests)
endif
	PIGEON_RELAY_HOST=carrier-pigeon.fly.dev PIGEON_RELAY_PORT=4433 PIGEON_TOKEN=$(PIGEON_TOKEN) \
		swift run pigeon-e2e-swift

# --- Benchmarks ---

bench:
	go test -bench=. -benchtime=2s -count=1 -timeout=120s -run=^$$ .

bench-live:
ifndef PIGEON_TOKEN
	$(error PIGEON_TOKEN is required for live benchmarks)
endif
	PIGEON_TOKEN=$(PIGEON_TOKEN) go test -bench=. -benchtime=2s -count=1 -timeout=120s -run=^$$ .

# --- Server ---

server:
	go run ./cmd/pigeon

# --- Code generation ---

generate:
	go run ./cmd/protogen protocol/pairing.yaml

# --- C library ---

amalgamate: generate
	./c/amalgamate.sh dist

test-c: amalgamate
	clang -DPIGEON_CRYPTO_LIBSODIUM -Idist $$(pkg-config --cflags --libs libsodium) \
		dist/pigeon.c c/test/test_pigeon.c -o c/test/test_pigeon
	./c/test/test_pigeon

# --- Clean ---

clean:
	rm -rf .build/
	rm -f pigeon pigeon-test-binary pigeon-e2e-server c/test/test_pigeon
	go clean -testcache
