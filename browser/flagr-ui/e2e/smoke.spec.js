import { test, expect } from '@playwright/test'

test.describe('App shell', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('page loads and displays the app shell', async ({ page }) => {
    await expect(page.locator('.logo')).toHaveText('Flagr')
    await expect(page).toHaveTitle('Flagr')
    await expect(page.locator('.navbar a[href*="api_docs"]')).toBeVisible()
  })

  test('navigation links are present', async ({ page }) => {
    await expect(page.locator('a[href*="api_docs"]')).toBeVisible()
    await expect(page.locator('a[href*="flagr"][target="_blank"]').first()).toBeVisible()
  })
})
