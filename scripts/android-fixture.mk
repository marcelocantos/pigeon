# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0
#
# Isolated clean-gate recipe for the Android device E2E Go fixture.
# The fixture links the pinned vendored libsodium archive, which is absent
# from a fresh checkout until its submodule and host library are built.
.PHONY: test-go

test-go:
	git submodule update --init c/vendor/github.com/jedisct1/libsodium
	bash c/vendor/build.sh libsodium
	go test -count=1 ./cmd/pigeon-android-fixture/...
