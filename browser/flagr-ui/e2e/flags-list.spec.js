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

  test('create flag via UI modal', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('[data-testid="flags-table"]')).toBeVisible({ timeout: 10000 })

    const flagName = `e2e-test-${Date.now()}`

    // Click "Create Flag" button to open the modal
    await page.locator('[data-testid="create-flag-btn"]').click()
    await expect(page.locator('.el-dialog')).toBeVisible({ timeout: 5000 })

    // Fill the description input inside the modal
    await page.locator('input[data-testid="new-flag-input"]').fill(flagName)

    // Click "Create Flag" inside the dialog footer (not the dropdown arrow)
    await page.getByRole('dialog').getByRole('button', { name: 'Create Flag' }).click()

    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })

    // Verify the flag appears in the list
    await expect(page.locator('[data-testid="flags-table"]')).toContainText(flagName)

    // Capture the flag id for cleanup by reading it back from the API
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

    // Search input uses a dedicated data-testid
    const searchInput = page.locator('[data-testid="search-input"]')
    await searchInput.fill(flagName)

    await expect(page.locator('[data-testid="flags-table"]')).toContainText(flagName)
  })
})
