// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// SwiftPM CPigeon-target forwarder for the canonical loopback header.
// The implementation lives in c/src/loopback.c, the canonical header
// in c/include/pigeon/loopback.h, and the amalgamation emits
// dist/loopback.h alongside dist/pigeon.h. This file just forwards so
// `import CPigeon` in Swift code resolves the same symbols.

#ifndef PIGEON_LOOPBACK_FORWARD_H
#define PIGEON_LOOPBACK_FORWARD_H

// Distinct guard from PIGEON_LOOPBACK_H so this forwarder can be
// included before dist/loopback.h and let the canonical header's
// declarations through.
#include "../../../dist/loopback.h"

#endif // PIGEON_LOOPBACK_FORWARD_H
