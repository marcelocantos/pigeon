// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// CPigeon SwiftPM C-target — compiles the amalgamated libpigeon source
// (dist/pigeon.c). Together with the Pigeon Swift target this gives
// XCTest a complete C peer library plus the in-process loopback
// transport that the multi-channel session API uses for no-I/O round
// trips.
//
// dist/pigeon.c includes the loopback transport since the amalgamation
// bundles c/src/loopback.c. The public declarations come in via
// dist/loopback.h (forwarded through include/pigeon_loopback.h).
//
// Pulling dist/pigeon.c via #include rather than listing it as a
// SwiftPM source keeps the C amalgamation in /dist (its canonical
// location, also consumed by cwire and the standalone `make test-c`
// build) without SwiftPM complaining that sources live outside the
// target's path.

#include "include/pigeon_loopback.h"

// Compile the amalgamated peer library inline. This already contains
// the loopback transport (see c/amalgamate.sh).
#include "../../dist/pigeon.c"
