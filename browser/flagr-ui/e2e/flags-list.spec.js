import { test, expect } from '@playwright/test'
import { createFlag, deleteFlag, API } from './helpers.js'

test.describe('Flags list page', () => {
  let createdId = null

  test.afterEach(async () => {
    if (createdId) {
      await deleteFlag(createdId).catch(() => {})
    }
  })

  test('page loads and shows the flags table', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('[data-testid="flags-table"]')).toBeVisible({ timeout: 10000 })
  })

  test('create flag via UI', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('[data-testid="flags-table"]')).toBeVisible({ timeout: 10000 })

    const flagName = `e2e-test-${Date.now()}`
    await page.locator('input[data-testid="new-flag-input"]').fill(flagName)
    await page.locator('button:has-text("Create New Flag")').click()

    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })

    // Verify the flag appears in the list
    await expect(page.locator('[data-testid="flags-table"]')).toContainText(flagName)

    // Capture the flag id for cleanup by reading it back from the API
    // so afterEach can clean it up even if this test doesn't navigates away
    const r = await page.request.get(`${API}/flags?description=${encodeURIComponent(flagName)}`)
    if (r.ok()) {
      const list = await r.json()
      if (list.length > 0) createdId = list[0].id
    }
  })

  test('search filters the flags list', async ({ page }) => {
    const flagName = `search-flag-${Date.now()}`
    const createResp = await page.request.post(`${API}/flags`, {
      data: { description: flagName }
    })
    expect(createResp.ok()).toBeTruthy()
    const created = await createResp.json()
    createdId = created.id

    await page.goto('/')
    await expect(page.locator('[data-testid="flags-table"]')).toBeVisible({ timeout: 10000 })

    const searchInput = page.locator('input[placeholder="Search a flag"]')
    await searchInput.fill(flagName)

    await expect(page.locator('[data-testid="flags-table"]')).toContainText(flagName)
  })
})
