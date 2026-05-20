// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

// This file used to expose the Go-side SessionMachine state via
// MachineState() for the T39.3 executor showcase test. T34c.3d retired
// the Go-side machine — activation now runs in libpigeon via cwire and
// the post-activation machine state is no longer tracked on the Go
// side. The file is kept for any future package-private hooks tests
// may need.
