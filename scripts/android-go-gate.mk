# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0
# Additional prerequisite for the pristine Android fixture Go gate.
# Keep Makefile's real test-go recipe: this only supplies its pinned native dep.
test-go: c/vendor/build/lib/libsodium.a
