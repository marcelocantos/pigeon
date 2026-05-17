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

# Append session_gen.h (post-T39 SessionMachine declarations) and the
# activation driver header. Both protogen-generated headers can coexist
# in one TU now that PIGEON_<protocol>_<KIND>_<NAME> per-protocol
# prefixes resolve the namespace collision (T32.1).
#
# These are appended OUTSIDE pigeon.h's PIGEON_H header guard, so wrap
# them in their own guard to keep re-inclusion (e.g. from loopback.h
# re-including pigeon.h) idempotent.
{
    echo ""
    echo "#ifndef PIGEON_H_AMALGAMATED_EXTRAS"
    echo "#define PIGEON_H_AMALGAMATED_EXTRAS"
    echo ""
    echo "// --- SessionMachine declarations (from session_gen.h) ---"
    echo ""
    sed -e '/#ifndef PIGEON_SESSION_GEN_H/d' \
        -e '/#define PIGEON_SESSION_GEN_H/d' \
        -e '/#endif.*PIGEON_SESSION_GEN_H/d' \
        -e '/#include <stdbool.h>/d' \
        -e '/#include <stdint.h>/d' \
        -e '/^\/\/ Copyright/d' \
        -e '/^\/\/ SPDX/d' \
        -e '/^\/\/ Code generated/d' \
        "$SRCDIR/include/pigeon/session_gen.h"
    echo ""
    echo "// --- Activation handshake driver (from activation.h) ---"
    echo ""
    sed -e '/#ifndef PIGEON_ACTIVATION_H/d' \
        -e '/#define PIGEON_ACTIVATION_H/d' \
        -e '/#endif.*PIGEON_ACTIVATION_H/d' \
        -e '/#include <stdbool.h>/d' \
        -e '/#include <stddef.h>/d' \
        -e '/#include <stdint.h>/d' \
        -e '/#include "pigeon.h"/d' \
        -e '/#include "session_gen.h"/d' \
        -e '/^\/\/ Copyright/d' \
        -e '/^\/\/ SPDX/d' \
        "$SRCDIR/include/pigeon/activation.h"
    echo ""
    echo "#endif // PIGEON_H_AMALGAMATED_EXTRAS"
} >> "$OUTDIR/pigeon.h"

# Also emit dist/loopback.h alongside dist/pigeon.h so language wrappers
# (cwire, SwiftPM CPigeon, JNI shim) can pull in the in-process test
# harness via a single extra include.
sed -e '/^\/\/ Copyright/d' \
    -e '/^\/\/ SPDX/d' \
    -e 's:#include "pigeon.h":#include "pigeon.h":' \
    "$SRCDIR/include/pigeon/loopback.h" \
    > "$OUTDIR/loopback.h"

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

    # Generated state machine implementations (strip includes + copyright).
    echo ""
    echo "// --- Generated pairing-ceremony state machine ---"
    echo ""
    sed -e '/^#include/d' \
        -e '/^\/\/ Copyright/d' \
        -e '/^\/\/ SPDX/d' \
        -e '/^\/\/ Code generated/d' \
        "$SRCDIR/src/pairingceremony_gen.c"

    echo ""
    echo "// --- Generated session state machine ---"
    echo ""
    sed -e '/^#include/d' \
        -e '/^\/\/ Copyright/d' \
        -e '/^\/\/ SPDX/d' \
        -e '/^\/\/ Code generated/d' \
        "$SRCDIR/src/session_gen.c"

    echo ""
    echo "// --- Activation handshake driver ---"
    echo ""
    sed -e '/^#include/d' \
        -e '/^\/\/ Copyright/d' \
        -e '/^\/\/ SPDX/d' \
        "$SRCDIR/src/activation.c"

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

    # Pairing ceremony wire driver (strip includes + copyright).
    echo ""
    echo "// --- Pairing ceremony driver ---"
    echo ""
    sed -e '/^#include/d' \
        -e '/^\/\/ Copyright/d' \
        -e '/^\/\/ SPDX/d' \
        "$SRCDIR/src/pairing.c"

    # Multi-client Listener (T32.2). Pull in after activation +
    # session_gen so the activation driver and the SessionMachine
    # state constants are visible at this point in the TU.
    echo ""
    echo "// --- Multi-client listener ---"
    echo ""
    sed -e '/^#include/d' \
        -e '/^\/\/ Copyright/d' \
        -e '/^\/\/ SPDX/d' \
        "$SRCDIR/src/listener.c"

    # Loopback transport (rewrite the header include from
    # "pigeon/loopback.h" to "loopback.h" so dist/loopback.h resolves
    # via the consumer's -I flag, strip system + copyright lines).
    echo ""
    echo "// --- In-process loopback transport ---"
    echo ""
    sed -e 's:#include "pigeon/loopback.h":#include "loopback.h":' \
        -e '/^#include <stdlib.h>/d' \
        -e '/^#include <string.h>/d' \
        -e '/^\/\/ Copyright/d' \
        -e '/^\/\/ SPDX/d' \
        "$SRCDIR/src/loopback.c"

} > "$OUTDIR/pigeon.c"

echo "wrote $OUTDIR/pigeon.h"
echo "wrote $OUTDIR/pigeon.c"
