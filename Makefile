# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0

JDK21 ?= /opt/homebrew/Cellar/openjdk@21/21.0.10/libexec/openjdk.jdk/Contents/Home

.PHONY: all build test test-go test-swift test-kotlin test-web \
        e2e e2e-go e2e-swift e2e-kotlin \
        test-live bench clean \
        build-vendor-deps test-c test-c-ngtcp2

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

# --- Deploy to Fly.io ---

deploy:
	flyctl deploy

# --- E2E tests against live relay (require PIGEON_TOKEN) ---

e2e-live: deploy e2e-go-live e2e-swift-live

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

# --- Vendored C dependencies (ngtcp2 + quictls/openssl) ---
#
# Builds static libs under vendor/build/.
# Outputs:
#   vendor/build/lib/libssl.a
#   vendor/build/lib/libcrypto.a
#   vendor/build/lib/libngtcp2.a
#   vendor/build/lib/libngtcp2_crypto_quictls.a
#
# Requires: cmake, make, perl (for OpenSSL Configure).
# Run once; subsequent builds skip if the sentinel file exists.

VENDOR_SENTINEL = c/vendor/build/lib/libngtcp2_crypto_quictls.a

build-vendor-deps: $(VENDOR_SENTINEL)

$(VENDOR_SENTINEL):
	bash c/vendor/build.sh all

# --- ngtcp2 QUIC transport test ---
#
# Links against the vendored static libraries built by build-vendor-deps.
# Unit tests (vtable wiring, struct layout) run without a live relay.
# Integration test skips gracefully if no relay is running on :4433.

VENDOR_BUILD = c/vendor/build
NGTCP2_CFLAGS = \
	-I$(VENDOR_BUILD)/include \
	-Ic/include \
	-Idist

NGTCP2_LDFLAGS = \
	$(VENDOR_BUILD)/lib/libngtcp2_crypto_quictls.a \
	$(VENDOR_BUILD)/lib/libngtcp2.a \
	$(VENDOR_BUILD)/lib/libssl.a \
	$(VENDOR_BUILD)/lib/libcrypto.a \
	-lpthread -ldl

test-c-ngtcp2: build-vendor-deps amalgamate
	clang $(NGTCP2_CFLAGS) \
		c/src/ngtcp2_transport.c \
		c/test/test_ngtcp2.c \
		dist/pigeon.c \
		-DPIGEON_CRYPTO_LIBSODIUM \
		$$(pkg-config --cflags --libs libsodium) \
		$(NGTCP2_LDFLAGS) \
		-o c/test/test_ngtcp2
	./c/test/test_ngtcp2

# --- Clean ---

clean:
	rm -rf .build/
	rm -f pigeon pigeon-test-binary pigeon-e2e-server c/test/test_pigeon c/test/test_ngtcp2
	go clean -testcache
