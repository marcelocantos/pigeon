// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Fixed port for the local relay's WebTransport (HTTP/3) listener in
// unattended E2E test mode. This must be a known constant at Playwright
// config load time so --origin-to-force-quic-on can reference it.
export const LOCAL_WT_PORT = 14433;
