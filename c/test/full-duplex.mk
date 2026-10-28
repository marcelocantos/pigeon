# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0
# Isolated clean-checkout test-go gate for the C SDK. The ordinary Makefile
# cannot build inside bin/gate -clean until its git submodules are populated.
.PHONY: test-go

test-go:
	git submodule update --init --recursive -- \
		c/vendor/github.com/jedisct1/libsodium \
		c/vendor/github.com/ngtcp2/ngtcp2 \
		c/vendor/github.com/quictls/openssl
	$(MAKE) build-vendor-deps
	bash c/amalgamate.sh dist
	$(MAKE) test-c-ngtcp2
	go test -a -tags csdke2e -count=1 -timeout=120s -v ./...
