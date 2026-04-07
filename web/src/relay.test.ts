// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

import { describe, it } from "node:test";

// relay.ts uses the browser-native WebTransport API which is not available in
// Node.js. E2E relay tests are covered by relay.e2e.ts (via pigeon-bridge).
// Unit tests require a WebTransport mock — none exists for Node.js yet.
describe("relay", () => {
  it.todo("needs WebTransport mock for unit testing");
});
