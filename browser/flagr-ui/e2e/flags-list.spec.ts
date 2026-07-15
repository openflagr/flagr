import { test, expect } from '@playwright/test'
import { createFlag, deleteFlag, getSnapshotMaxId, API } from './helpers'

test.describe('Flags list page', () => {
  let createdId = null

  test.afterEach(async () => {
    if (createdId) {
      await deleteFlag(createdId).catch(() => {})
    }
  })

  test('page loads and shows the flags list or empty state', async ({ page }) => {
    await page.goto('/')
    await expect(
      page.locator('[data-testid="flags-table"], [data-testid="flags-empty"]'),
    ).toBeVisible({ timeout: 10000 })
  })

  test('create flag via UI modal', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('[data-testid="create-flag-btn"]')).toBeVisible({ timeout: 10000 })

    const flagName = `e2e-test-${Date.now()}`

    // Click "Create Flag" button to open the modal
    await page.locator('[data-testid="create-flag-btn"]').click()
    await expect(page.locator('.el-dialog')).toBeVisible({ timeout: 5000 })

    // Fill the description input inside the modal
    await page.locator('input[data-testid="new-flag-input"]').fill(flagName)

    // Click "Create Flag" inside the dialog footer (not the dropdown arrow)
    await page.getByRole('dialog').getByRole('button', { name: 'Create Flag' }).click()

    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })

    // Verify the flag appears exactly once in the list (no duplicates)
    const matchingRows = page.locator('[data-testid="flags-table"] .el-table__row', { hasText: flagName })
    await expect(matchingRows).toHaveCount(1)

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

test.describe('Flag list cache & snapshot max-id', () => {
  let createdId = null

  test.afterEach(async () => {
    if (createdId) {
      await deleteFlag(createdId).catch(() => {})
    }
  })

  test('GET /flags/snapshots/max_id returns a non-negative integer', async () => {
    const maxId = await getSnapshotMaxId()
    expect(typeof maxId).toBe('number')
    expect(maxId).toBeGreaterThanOrEqual(0)
    expect(Number.isInteger(maxId)).toBe(true)
  })

  test('max-id increases after creating a flag', async () => {
    const before = await getSnapshotMaxId()
    const flag = await createFlag()
    createdId = flag.id

    // Wait for snapshot to be created
    await new Promise(r => setTimeout(r, 500))

    const after = await getSnapshotMaxId()
    expect(after).toBeGreaterThan(before)
  })

  test('flag list loads and caches correctly across navigations', async ({ page }) => {
    // Create a flag first
    const flag = await createFlag()
    createdId = flag.id
    const desc = flag.description

    // First visit — loads flags via API
    await page.goto('/')
    await expect(page.locator('[data-testid="flags-table"]')).toContainText(desc, { timeout: 10000 })

    // Open the flag we created (do not use .first() — parallel workers add other rows)
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })

    // Navigate back — should use cached flags
    await page.goto('/')
    await expect(page.locator('[data-testid="flags-table"]')).toContainText(desc, { timeout: 10000 })
  })
})
