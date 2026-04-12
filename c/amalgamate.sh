#!/bin/sh
# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0
#
# Amalgamate the pigeon C library into a single pigeon.h / pigeon.c pair.
# Usage: ./c/amalgamate.sh [outdir]

set -e

OUTDIR="${1:-dist}"
SRCDIR="$(cd "$(dirname "$0")" && pwd)"

mkdir -p "$OUTDIR"

# --- pigeon.h ---
# Inline the generated header into pigeon.h, replacing the #include directive.
# Extract generated header content (strip guards, duplicate system includes,
# copyright, and blank lines at boundaries).
cat "$SRCDIR/include/pigeon/pairingceremony_gen.h" \
    | sed '/#ifndef PIGEON_PAIRINGCEREMONY_GEN_H/d' \
    | sed '/#define PIGEON_PAIRINGCEREMONY_GEN_H/d' \
    | sed '/#endif.*PIGEON_PAIRINGCEREMONY_GEN_H/d' \
    | sed '/#include <stdbool.h>/d' \
    | sed '/#include <stdint.h>/d' \
    | sed '/^\/\/ Copyright/d' \
    | sed '/^\/\/ SPDX/d' \
    | sed '/^\/\/ Code generated/d' \
    | awk 'NF{p=1} p' \
    > "$OUTDIR/.gen_fragment.h"

# Replace the #include directive with the fragment content, then clean up
# the stale comment above it.
sed '/#include "pairingceremony_gen.h"/r '"$OUTDIR/.gen_fragment.h" \
    "$SRCDIR/include/pigeon/pigeon.h" \
    | sed '/#include "pairingceremony_gen.h"/d' \
    | sed '/^\/\/ Include the generated protocol header\.$/d' \
    > "$OUTDIR/pigeon.h"

rm -f "$OUTDIR/.gen_fragment.h"

# --- pigeon.c ---
{
    cat <<'HEADER'
// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// Pigeon C client library — amalgamated source.
// Compile with -DPIGEON_CRYPTO_LIBSODIUM and link -lsodium.

#include "pigeon.h"
#include <string.h>
HEADER

    # Generated state machine implementation (strip includes + copyright).
    echo ""
    echo "// --- Generated state machine ---"
    echo ""
    sed -e '/^#include/d' \
        -e '/^\/\/ Copyright/d' \
        -e '/^\/\/ SPDX/d' \
        -e '/^\/\/ Code generated/d' \
        "$SRCDIR/src/pairingceremony_gen.c"

    # Crypto implementation (strip includes + copyright, keep #if guards).
    echo ""
    echo "// --- Crypto ---"
    echo ""
    sed -e '/^#include "pigeon\/pigeon.h"/d' \
        -e '/^#include <string.h>/d' \
        -e '/^\/\/ Copyright/d' \
        -e '/^\/\/ SPDX/d' \
        "$SRCDIR/src/crypto.c"

    # Conn/framing implementation (strip includes + copyright).
    echo ""
    echo "// --- Connection and framing ---"
    echo ""
    sed -e '/^#include/d' \
        -e '/^\/\/ Copyright/d' \
        -e '/^\/\/ SPDX/d' \
        "$SRCDIR/src/pigeon.c"

} > "$OUTDIR/pigeon.c"

echo "wrote $OUTDIR/pigeon.h"
echo "wrote $OUTDIR/pigeon.c"
