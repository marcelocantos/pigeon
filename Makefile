# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0

JDK21 ?= /opt/homebrew/Cellar/openjdk@21/21.0.11/libexec/openjdk.jdk/Contents/Home

.PHONY: all build test test-go test-swift test-kotlin test-web \
        e2e e2e-go e2e-swift e2e-kotlin \
        test-live bench clean \
        build-vendor-deps test-c test-c-only test-c-asan test-c-ngtcp2 \
        test-go-race \
        bullseye bullseye-prereq bullseye-strict demo server

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
	cd web && npx tsx --test src/crypto.test.ts src/PairingCeremonyMachine.test.ts src/relay.test.ts

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

# --- Demo ---
#
# Runs the full pigeon stack — relay, backend, client — in one process
# behind an HTTP control server, and opens a web UI at
# http://127.0.0.1:7000 showing live traffic across every channel.

demo:
	go run ./examples/demo

# --- Code generation ---

generate:
	go run ./cmd/protogen protocol/pairing.yaml
	go run ./cmd/protogen --wireformats protocol/wireformats.yaml

# --- C library ---

VENDOR_BUILD = c/vendor/build
SODIUM_SENTINEL = $(VENDOR_BUILD)/lib/libsodium.a
SODIUM_CFLAGS = -I$(VENDOR_BUILD)/include
SODIUM_LDFLAGS = $(SODIUM_SENTINEL)

amalgamate: generate
	./c/amalgamate.sh dist

test-c: amalgamate test-c-only

# test-c without the amalgamate prereq, for callers (bullseye) that
# already serialised generate+amalgamate up front and must not re-run
# them in a parallel branch — the regen would rewrite Swift sources
# mid-compile and trip "input file was modified during the build".
test-c-only: $(SODIUM_SENTINEL)
	clang -DPIGEON_CRYPTO_LIBSODIUM -Idist $(SODIUM_CFLAGS) \
		dist/pigeon.c c/test/test_pigeon.c $(SODIUM_LDFLAGS) \
		-o c/test/test_pigeon
	./c/test/test_pigeon

# --- Sanitiser-instrumented C tests ---
#
# AddressSanitizer + UndefinedBehaviorSanitizer catch the kinds of
# memory bugs the heap-scratch and cgo-bridge work introduces (use-
# after-free, leaks, oob, signed overflow, null deref). Run as a
# separate target so the bullseye fast loop stays fast.
#
# UBSan settles for -fno-sanitize=function on Apple silicon where the
# function-type check trips on libsodium's call-into-C pattern.
test-c-asan: amalgamate $(SODIUM_SENTINEL)
	clang -O1 -g -fno-omit-frame-pointer \
		-fsanitize=address,undefined -fno-sanitize=function \
		-DPIGEON_CRYPTO_LIBSODIUM -Idist $(SODIUM_CFLAGS) \
		dist/pigeon.c c/test/test_pigeon.c $(SODIUM_LDFLAGS) \
		-o c/test/test_pigeon_asan
	ASAN_OPTIONS=detect_leaks=1:abort_on_error=1:halt_on_error=1 \
		UBSAN_OPTIONS=halt_on_error=1:print_stacktrace=1 \
		./c/test/test_pigeon_asan

# Go race detector on the cwire bridge (covers AEAD nonce races etc).
test-go-race:
	go test -race -count=1 -timeout=120s ./cwire/ ./crypto/ ./

# --- Vendored C dependencies (libsodium + ngtcp2 + quictls/openssl) ---
#
# Builds static libs under vendor/build/. libsodium is split out
# because test-c only needs it (not openssl/ngtcp2), so a fast loop
# can build libsodium alone (~10s) rather than the whole stack.
# Outputs:
#   vendor/build/lib/libsodium.a
#   vendor/build/lib/libssl.a
#   vendor/build/lib/libcrypto.a
#   vendor/build/lib/libngtcp2.a
#   vendor/build/lib/libngtcp2_crypto_quictls.a
#
# Requires: cmake, make, perl (for OpenSSL Configure), autoconf +
# automake + libtool (for libsodium's autogen.sh).
# Run once; subsequent builds skip if the sentinel file exists.

VENDOR_SENTINEL = $(VENDOR_BUILD)/lib/libngtcp2_crypto_quictls.a

build-vendor-deps: $(SODIUM_SENTINEL) $(VENDOR_SENTINEL)

$(SODIUM_SENTINEL):
	bash c/vendor/build.sh libsodium

$(VENDOR_SENTINEL):
	bash c/vendor/build.sh openssl
	bash c/vendor/build.sh ngtcp2

# --- ngtcp2 QUIC transport test ---
#
# Links against the vendored static libraries built by build-vendor-deps.
# Unit tests (vtable wiring, struct layout) run without a live relay.
# Integration test skips gracefully if no relay is running on :4433.

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
	clang $(NGTCP2_CFLAGS) $(SODIUM_CFLAGS) \
		c/src/ngtcp2_transport.c \
		c/test/test_ngtcp2.c \
		dist/pigeon.c \
		-DPIGEON_CRYPTO_LIBSODIUM \
		$(NGTCP2_LDFLAGS) $(SODIUM_LDFLAGS) \
		-o c/test/test_ngtcp2
	./c/test/test_ngtcp2
	@# 🎯T32.4: cgo-driven live round-trips through the C SDK
	@# (pigeon_register / pigeon_connect / pigeon_listener_accept)
	@# against an in-process Go relay (pigeon.NewQUICServer). Build
	@# tag csdke2e keeps the package out of the default Go build —
	@# only callers with the vendored ngtcp2 + quictls static libs
	@# already built (which `build-vendor-deps` above guarantees)
	@# can link it.
	go test -tags csdke2e -count=1 -timeout=120s ./c/test/csdke2e/

# --- Standing invariants (for bullseye_convergence) ---

# Bullseye runs the SDK suites and TLC concurrently. Each step writes
# its own log and a tiny status file (.ok or .fail-<step>); the main
# target waits, then renders the per-step pass/fail summary in stable
# order. Total wall-clock drops from ~serial-sum to roughly the slowest
# single suite (today: swift test).
#
# `make bullseye-strict` adds ASan + UBSan on the C suite and Go's
# race detector on cwire / crypto / root. Slower (~30-40s vs ~10s for
# bullseye) so it's a separate target rather than always-on.
bullseye-strict: bullseye test-c-asan test-go-race
	@echo "✓ bullseye-strict (sanitisers green)"

# Sequential prerequisites for bullseye: codegen and amalgamation must
# happen before the parallel checks fan out, otherwise `make generate`
# / `make amalgamate` (transitively required by test-c) races against
# the gofmt and go vet/build/test branches scanning the same files.
# Also build libsodium up front so the test-c branch doesn't race
# against itself if multiple bullseye-style targets reach it.
bullseye-prereq: $(SODIUM_SENTINEL)
	@$(MAKE) -s amalgamate >/dev/null

bullseye: bullseye-prereq
	@rm -rf /tmp/bullseye && mkdir -p /tmp/bullseye
	@( out=$$(gofmt -l .); \
	   if test -z "$$out"; then echo ok > /tmp/bullseye/gofmt.status; \
	   else { echo "$$out" > /tmp/bullseye/gofmt.log; echo fail > /tmp/bullseye/gofmt.status; }; fi ) & \
	 ( go vet ./... > /tmp/bullseye/govet.log 2>&1 \
	   && echo ok > /tmp/bullseye/govet.status \
	   || echo fail > /tmp/bullseye/govet.status ) & \
	 ( go build ./... > /tmp/bullseye/gobuild.log 2>&1 \
	   && echo ok > /tmp/bullseye/gobuild.status \
	   || echo fail > /tmp/bullseye/gobuild.status ) & \
	 ( go test -count=1 -short -timeout=300s ./... > /tmp/bullseye/gotest.log 2>&1 \
	   && echo ok > /tmp/bullseye/gotest.status \
	   || echo fail > /tmp/bullseye/gotest.status ) & \
	 ( swift test > /tmp/bullseye/swift-test.log 2>&1 \
	   && echo ok > /tmp/bullseye/swift-test.status \
	   || echo fail > /tmp/bullseye/swift-test.status ) & \
	 ( JAVA_HOME=$(JDK21) android/gradlew -p $(CURDIR)/android :pigeon:test \
	     --no-daemon --console=plain > /tmp/bullseye/kotlin.log 2>&1 \
	   && echo ok > /tmp/bullseye/kotlin.status \
	   || echo fail > /tmp/bullseye/kotlin.status ) & \
	 ( cd web && npx tsx --test src/crypto.test.ts src/PairingCeremonyMachine.test.ts src/relay.test.ts \
	     > /tmp/bullseye/web.log 2>&1 \
	   && echo ok > /tmp/bullseye/web.status \
	   || echo fail > /tmp/bullseye/web.status ) & \
	 ( $(MAKE) test-c-only > /tmp/bullseye/test-c.log 2>&1 \
	   && echo ok > /tmp/bullseye/test-c.status \
	   || echo fail > /tmp/bullseye/test-c.status ) & \
	 ( cd formal && ./tlc PairingCeremony > /tmp/bullseye/tlc.log 2>&1 \
	     && ./tlc SessionMachine >> /tmp/bullseye/tlc.log 2>&1; \
	   completed=$$(grep -c "Model checking completed. No error has been found" /tmp/bullseye/tlc.log); \
	   if [ "$$completed" = "2" ]; then \
	     echo ok > /tmp/bullseye/tlc.status; \
	   else \
	     echo fail > /tmp/bullseye/tlc.status; \
	   fi ) & \
	 wait
	@fail=0; \
	 for step in gofmt govet gobuild gotest swift-test kotlin web test-c tlc; do \
	   if [ "$$(cat /tmp/bullseye/$$step.status 2>/dev/null)" = "ok" ]; then \
	     printf "✓ %s\n" "$$step"; \
	   else \
	     printf "✗ %s\n" "$$step"; \
	     tail -30 /tmp/bullseye/$$step.log 2>/dev/null | sed 's/^/    /'; \
	     fail=1; \
	   fi; \
	 done; \
	 exit $$fail

# --- Clean ---

clean:
	rm -rf .build/
	rm -f pigeon pigeon-test-binary pigeon-e2e-server c/test/test_pigeon c/test/test_ngtcp2
	go clean -testcache
