import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  timeout: 30000,
  expect: {
    timeout: 10000
  },
  fullyParallel: false,
  retries: process.env.CI ? 2 : 0,
  reporter: process.env.CI
    ? [['html'], ['json', { outputFile: 'results.json' }]]
    : 'list',
  use: {
    baseURL: 'http://localhost:8080',
    headless: true,
    actionTimeout: 10000,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  webServer: {
    command: 'npm run serve',
    url: 'http://localhost:8080',
    reuseExistingServer: !process.env.CI,
    timeout: 120000,
  },
})
