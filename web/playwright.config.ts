// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  timeout: 30_000,
  retries: 0,
  use: {
    // Chromium is the only browser with WebTransport support.
    browserName: "chromium",
    headless: true,
  },
  projects: [
    {
      name: "chromium",
      use: {
        browserName: "chromium",
        launchOptions: {
          // Chromium treats localhost as a secure context, so no cert
          // flags are needed for the test page. WebTransport to
          // tern.fly.dev uses a real Let's Encrypt cert.
        },
      },
    },
  ],
});
