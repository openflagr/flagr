import { test, expect } from '@playwright/test'
import { createFlag, deleteFlag } from './helpers.js'

test.describe('Flags list page', () => {
  test('page loads and shows the flags table', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('[data-testid="flags-table"]')).toBeVisible({ timeout: 10000 })
  })

  test('create flag via UI', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('[data-testid="flags-table"]', { timeout: 10000 })

    const flagName = `e2e-test-${Date.now()}`
    await page.locator('input[data-testid="new-flag-input"]').fill(flagName)
    await page.locator('button:has-text("Create New Flag")').click()

    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })

    // Verify the flag appears in the list
    await expect(page.locator('[data-testid="flags-table"]')).toContainText(flagName)
  })

  test('search filters the flags list', async ({ page }) => {
    const flagName = `search-flag-${Date.now()}`
    const createResp = await page.request.post('http://localhost:18000/api/v1/flags', {
      data: { description: flagName }
    })
    expect(createResp.ok()).toBeTruthy()
    const created = await createResp.json()

    await page.goto('/')
    await page.waitForSelector('[data-testid="flags-table"]', { timeout: 10000 })

    const searchInput = page.locator('input[placeholder="Search a flag"]')
    await searchInput.fill(flagName)

    await expect(page.locator('[data-testid="flags-table"]')).toContainText(flagName)

    await deleteFlag(created.id)
  })
})
