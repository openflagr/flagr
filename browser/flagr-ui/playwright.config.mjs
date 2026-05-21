import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: '.',
  timeout: 30000,
  retries: 0,
  use: {
    baseURL: 'http://localhost:8080',
    headless: true,
  },
  webServer: {
    command: 'npm run dev:full',
    port: 8080,
    reuseExistingServer: !process.env.CI,
    timeout: 30000,
  },
})
