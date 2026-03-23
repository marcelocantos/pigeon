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
          args: [
            // Playwright's bundled Chromium may not trust system CAs.
            // Allow connections to tern.fly.dev with any cert.
            "--ignore-certificate-errors",
          ],
        },
      },
    },
  ],
});
