import { defineConfig, devices } from '@playwright/test'

const workers = Math.min(4, Math.max(1, Number(process.env.PW_WORKERS) || 4))

/**
 * E2e isolation:
 * - One shared Flagr API + SQLite; tests must use unique flag descriptions (Date.now()) and per-test cleanup.
 * - Playwright runs one test at a time per worker; module-level `let flag` in a spec file is not shared across workers.
 * - Do not rely on flags table sort order; open detail via /#/flags/:id when the test created the flag.
 */
export default defineConfig({
  testDir: 'e2e',
  timeout: 30000,
  retries: process.env.CI ? 1 : 0,
  fullyParallel: true,
  workers,
  use: {
    baseURL: 'http://localhost:8080',
    headless: true,
    ...devices['Desktop Chrome'],
  },
  webServer: [
    {
      command: 'sh scripts/e2e-server.sh',
      port: 8080,
      reuseExistingServer: true,
      timeout: 60000,
    },
  ],
})