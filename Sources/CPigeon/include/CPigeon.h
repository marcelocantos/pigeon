// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

#ifndef CPIGEON_H
#define CPIGEON_H

// Umbrella header for the CPigeon SwiftPM C-target. Pulls in the
// amalgamated peer-library API (dist/pigeon.h) so Swift code can
// `import CPigeon` and call into pigeon_session_*, pigeon_stream_*,
// pigeon_datagram_*, the wire helpers, the pairing-ceremony FSM, and
// the AEAD channel primitives.
//
// This is part of T29 — the Swift peer library is being refactored
// from a Network.framework single-stream relay client into an
// idiomatic-Swift wrapper over libpigeon, mirroring the Go
// reference bridge in cwire/cwire.go.

#include "pigeon.h"
#include "pigeon_loopback.h"

#endif // CPIGEON_H
