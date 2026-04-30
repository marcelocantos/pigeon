// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// Wraps the amalgamated C library (dist/pigeon.c) into the cgo build
// of package cwire. We compile the amalgamation here rather than
// linking a static lib because libpigeon hasn't been packaged yet —
// once T32 step 3 finishes the multi-stream port, this file can be
// replaced with a `#cgo LDFLAGS: -lpigeon` and a binary distribution.
//
// dist/pigeon.c must be regenerated whenever the C sources change:
//   make amalgamate

#include "pigeon.c"
