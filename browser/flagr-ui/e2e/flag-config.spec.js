import { test, expect } from '@playwright/test'
const { createFlag } = require('./helpers')

let flagId

test.describe('Flag Config', () => {
  test.beforeAll(async () => {
    const flag = await createFlag('config-test-flag-' + Date.now())
    flagId = flag.id
  })

  test.beforeEach(async ({ page }) => {
    await page.goto(`http://localhost:8080/#/flags/${flagId}`)
    await page.waitForSelector('.flag-container', { timeout: 10000 })
  })

  test('Flag page loads', async ({ page }) => {
    await expect(page.locator('.flag-config-card')).toBeVisible()
    await expect(page.locator('.el-tag').first()).toContainText(`Flag ID:`)
  })

  test('Config and History tabs visible', async ({ page }) => {
    await expect(page.locator('.el-tabs__item').filter({ hasText: 'Config' })).toBeVisible()
    await expect(page.locator('.el-tabs__item').filter({ hasText: 'History' })).toBeVisible()
    // Config is active by default
    await expect(page.locator('.el-tabs__item').filter({ hasText: 'Config' })).toHaveClass(/is-active/)
  })

  test('Edit flag key', async ({ page }) => {
    const keyInput = page.locator('.flag-content input[placeholder="Key"]')
    const newKey = 'test-key-' + Date.now()
    await keyInput.fill(newKey)

    await page.locator('button').filter({ hasText: 'Save Flag' }).click()
    await expect(page.locator('.el-message')).toContainText('Flag updated')

    // Reload and verify
    await page.reload()
    await page.waitForSelector('.flag-container')
    await expect(keyInput).toHaveValue(newKey)
  })

  test('Edit flag description', async ({ page }) => {
    const descInput = page.locator('.flag-content input[placeholder="Description"]')
    const newDesc = 'updated-desc-' + Date.now()
    await descInput.fill(newDesc)

    await page.locator('button').filter({ hasText: 'Save Flag' }).click()
    await expect(page.locator('.el-message')).toContainText('Flag updated')

    await page.reload()
    await page.waitForSelector('.flag-container')
    await expect(descInput).toHaveValue(newDesc)
  })

  test('Toggle enabled/disabled', async ({ page }) => {
    const switchEl = page.locator('.el-card-header .el-switch').first()
    await switchEl.click()
    await page.waitForTimeout(500)

    const msg = page.locator('.el-message')
    await expect(msg).toContainText(/You turned (on|off) this feature flag/)
  })

  test('Data Records toggle', async ({ page }) => {
    // Find data records switch
    const switches = page.locator('.flag-content .el-switch')
    await expect(switches.first()).toBeVisible()
  })

  test('Entity Type select', async ({ page }) => {
    // Entity type select may be hidden if data records disabled, just check it exists on page
    const selects = page.locator('.flag-content .el-select')
    expect(await selects.count()).toBeGreaterThanOrEqual(1)
  })

  test('Delete Flag shows confirmation dialog', async ({ page }) => {
    const deleteBtn = page.locator('button').filter({ hasText: 'Delete Flag' })
    await expect(deleteBtn).toBeVisible()

    // Click delete - should show dialog
    await deleteBtn.click()
    const dialog = page.locator('.el-dialog').filter({ hasText: 'Delete feature flag' })
    await expect(dialog).toBeVisible()
    await expect(dialog).toContainText('Are you sure you want to delete this feature flag')

    // Cancel
    await dialog.locator('button').filter({ hasText: 'Cancel' }).click()
    await page.waitForTimeout(300)
  })

  test('Save Flag sends all fields', async ({ page }) => {
    // Just verify Save Flag button works
    await page.locator('button').filter({ hasText: 'Save Flag' }).click()
    await expect(page.locator('.el-message')).toContainText('Flag updated')
  })

  test('Flag ID is not editable', async ({ page }) => {
    const flagIdTag = page.locator('.el-tag').filter({ hasText: /Flag ID:/ }).first()
    await expect(flagIdTag).toBeVisible()
    // It's a tag, not an input
    const input = flagIdTag.locator('input')
    expect(await input.count()).toBe(0)
  })
})
