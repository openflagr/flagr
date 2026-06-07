import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: '.',
  timeout: 30000,
  retries: 0,
  use: {
    baseURL: 'http://localhost:8080',
    headless: true,
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
