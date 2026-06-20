// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

import { defineConfig } from "@playwright/test";
import { LOCAL_WT_PORT } from "./e2e/local-relay-port.js";

export default defineConfig({
  testDir: "./e2e",
  timeout: 60_000,
  retries: 0,
  // Start a local relay when PIGEON_TOKEN is not set (unattended mode).
  globalSetup: "./e2e/global-setup.ts",
  globalTeardown: "./e2e/global-setup.ts",
  use: {
    // Chromium is the only browser with WebTransport support.
    browserName: "chromium",
    // Use the full Chromium build (channel: "chromium") rather than the
    // default headless-shell. WebTransport/HTTP3 behaves identically, and
    // it sidesteps environments where only the full browser is installed.
    channel: "chromium",
    headless: true,
  },
  projects: [
    {
      name: "chromium",
      use: {
        browserName: "chromium",
        channel: "chromium",
        launchOptions: {
          args: [
            // Playwright's bundled Chromium may not trust system CAs.
            // Allow connections to carrier-pigeon.fly.dev with any cert.
            "--ignore-certificate-errors",
            // Force QUIC (and thus WebTransport) on the local relay port.
            // Required so Chromium uses HTTP/3 without a prior Alt-Svc hint.
            `--origin-to-force-quic-on=localhost:${LOCAL_WT_PORT}`,
          ],
        },
      },
    },
  ],
});
