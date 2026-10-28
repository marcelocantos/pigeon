#!/usr/bin/env bash
# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail
bash c/test/run-ngtcp2-fin.sh
# Force cgo to recompile its #included transport.c against this checkout.
go test -a -tags csdke2e -count=1 -timeout=120s ./c/test/csdke2e/
