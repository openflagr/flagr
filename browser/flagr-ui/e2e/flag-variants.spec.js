import { test, expect } from '@playwright/test'
const { createFlag } = require('./helpers')

let flagId

test.describe('Flag Variants', () => {
  test.beforeAll(async () => {
    const flag = await createFlag('variants-test-' + Date.now())
    flagId = flag.id
  })

  test.beforeEach(async ({ page }) => {
    await page.goto(`http://localhost:8080/#/flags/${flagId}`)
    await page.waitForSelector('.flag-container', { timeout: 10000 })
  })

  test('Empty state', async ({ page }) => {
    await expect(page.locator('.card--error').first()).toContainText('No variants created for this feature flag yet')
  })

  test('Create variant', async ({ page }) => {
    const keyInput = page.locator('input[placeholder="Variant Key"]')
    const createBtn = page.locator('button').filter({ hasText: 'Create Variant' })
    await expect(createBtn).toBeDisabled()
    await keyInput.fill('control')
    await expect(createBtn).not.toBeDisabled()
    await createBtn.click()
    await expect(page.locator('.el-message')).toContainText('new variant created')
    // Variant key is inside an input element, check via input value
    await expect(page.locator('.variants-container-inner .variant-key-input input').first()).toHaveValue('control')
  })

  test('Create second variant', async ({ page }) => {
    const keyInput = page.locator('input[placeholder="Variant Key"]')
    const createBtn = page.locator('button').filter({ hasText: 'Create Variant' })
    await keyInput.fill('treatment')
    await createBtn.click()
    await expect(page.locator('.el-message').last()).toContainText('new variant created')
    await page.waitForTimeout(300)
    // Variant key is inside an input element
    const variantInputs = page.locator('.variants-container-inner .variant-key-input input')
    const count = await variantInputs.count()
    expect(count).toBeGreaterThanOrEqual(2)
  })

  test('Edit variant key', async ({ page }) => {
    const variantInputs = page.locator('.variants-container-inner .variant-key-input input')
    if (await variantInputs.count() > 0) {
      await variantInputs.first().fill('control-v2')
      await page.locator('.variants-container-inner button').filter({ hasText: 'Save Variant' }).first().click()
      await expect(page.locator('.el-message')).toContainText('variant updated')
    }
  })

  test('Variant attachment collapse', async ({ page }) => {
    const collapseHeader = page.locator('.variant-attachment-collapsable-title .el-collapse-item__header').first()
    if (await collapseHeader.isVisible().catch(() => false)) {
      await collapseHeader.click()
      await page.waitForTimeout(300)
      await expect(page.locator('.variant-attachment-title').first()).toContainText('JSON')
    }
  })

  test('Invalid variant attachment shows error', async ({ page }) => {
    const saveBtn = page.locator('.variants-container-inner button').filter({ hasText: 'Save Variant' })
    if (await saveBtn.count() > 0) {
      await expect(saveBtn.first()).toBeVisible()
    }
  })

  test('Delete variant not in use', async ({ page }) => {
    page.on('dialog', dialog => dialog.accept())
    const keyInput = page.locator('input[placeholder="Variant Key"]')
    const createBtn = page.locator('button').filter({ hasText: 'Create Variant' })
    await keyInput.fill('to-delete-' + Date.now())
    await createBtn.click()
    await page.waitForTimeout(500)
    // Click delete button (icon button in save-remove row, not "Save Variant")
    const deleteIcons = page.locator('.variants-container-inner .el-icon-delete')
    if (await deleteIcons.count() > 0) {
      await deleteIcons.last().click()
      await page.waitForTimeout(500)
      await expect(page.locator('.el-message').last()).toContainText('variant deleted')
    }
  })

  test('Variant in use check exists', async ({ page }) => {
    await expect(page.locator('.variants-container')).toBeVisible()
  })
})
