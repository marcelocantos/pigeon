// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Playwright globalSetup: starts a local pigeon relay when PIGEON_TOKEN is
// not set, so the browser E2E tests can run unattended without a remote relay.
//
// The relay is started with --cert-validity 14 so its self-signed cert is
// accepted by Chromium via serverCertificateHashes (Chromium's 14-day limit).
// The cert hash and relay URL are passed to the test worker processes via
// environment variables:
//
//   PIGEON_RELAY_URL   — e.g. "https://localhost:14433"
//   PIGEON_CERT_HASH   — hex SHA-256 fingerprint (set only for local relay)
//
// The WebTransport port is fixed at LOCAL_WT_PORT (from playwright.config.ts)
// so that --origin-to-force-quic-on can be set statically in Chromium args.

import { startRelayOnPort } from "../src/test-helpers/GoRelayProcess.js";
import { LOCAL_WT_PORT } from "./local-relay-port.js";
import type { FullConfig } from "@playwright/test";

// Module-level handle so teardown can stop the relay.
let stopRelay: (() => void) | null = null;

export default async function globalSetup(_config: FullConfig): Promise<void> {
  // If a real relay token is provided, the tests connect to the live relay
  // and no local relay is needed.
  if (process.env.PIGEON_TOKEN) {
    return;
  }

  const relay = await startRelayOnPort(LOCAL_WT_PORT);
  stopRelay = relay.close;

  // Inject env vars that relay.spec.ts reads to find the relay and cert.
  process.env.PIGEON_RELAY_URL = `https://localhost:${relay.wtPort}`;
  process.env.PIGEON_CERT_HASH = relay.certHash;
  // Use a sentinel token so tests don't skip (local relay requires no token).
  process.env.PIGEON_TOKEN = "__local__";

  console.log(
    `[globalSetup] Local relay started on port ${relay.wtPort}, cert-hash=${relay.certHash}`,
  );
}

export async function globalTeardown(_config: FullConfig): Promise<void> {
  if (stopRelay) {
    stopRelay();
    stopRelay = null;
  }
}
