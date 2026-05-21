import { test, expect } from '@playwright/test'
import { createFlagWithVariants, createFlag, deleteFlag, createSegment, createConstraint, createVariant } from './helpers.js'

test.describe('Flag detail page', () => {
  test('loads and displays flag details', async ({ page }) => {
    const flag = await createFlagWithVariants()
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })
    await deleteFlag(flag.id)
  })

  test('shows all sections', async ({ page }) => {
    const flag = await createFlagWithVariants()
    await page.goto(`/#/flags/${flag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await expect(page.locator('h2:has-text("Variants")')).toBeVisible()
    await expect(page.locator('h2:has-text("Segments")').first()).toBeVisible()
    await expect(page.locator('h2:has-text("Flag Settings")')).toBeVisible()
    await deleteFlag(flag.id)
  })

  test('can toggle flag enabled state', async ({ page }) => {
    const flag = await createFlagWithVariants()
    await page.goto(`/#/flags/${flag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.locator('[data-testid="flag-enable-switch"]').click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    await deleteFlag(flag.id)
  })

  test('can update flag key and save', async ({ page }) => {
    const flag = await createFlagWithVariants()
    await page.goto(`/#/flags/${flag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.locator('input[data-testid="flag-key-input"]').fill(`key-${Date.now()}`)
    await page.locator('[data-testid="save-flag-btn"]').click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    await deleteFlag(flag.id)
  })

  test('can update variant key and save', async ({ page }) => {
    const flag = await createFlagWithVariants()
    await page.goto(`/#/flags/${flag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.locator('input[data-testid="variant-key-input"]').first().fill(`vk-${Date.now()}`)
    await page.locator('[data-testid="save-variant-btn"]').first().click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    await deleteFlag(flag.id)
  })

  test('can create a new variant', async ({ page }) => {
    const flag = await createFlagWithVariants()
    await page.goto(`/#/flags/${flag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.locator('input[data-testid="new-variant-input"]').fill(`v-${Date.now()}`)
    await page.locator('[data-testid="create-variant-btn"]').click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    await deleteFlag(flag.id)
  })

  test('can delete a variant', async ({ page }) => {
    const flag = await createFlagWithVariants()
    const delKey = `del-v-${Date.now()}`
    await createVariant(flag.id, delKey)
    await page.goto(`/#/flags/${flag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.locator('[data-testid="delete-variant-btn"]').last().click()
    await expect(page.locator('.el-message-box')).toBeVisible({ timeout: 5000 })
    await page.locator('.el-message-box .el-button--primary').click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    await deleteFlag(flag.id)
  })

  test('segment cards display correctly', async ({ page }) => {
    const flag = await createFlag()
    await createSegment(flag.id, `display-${Date.now()}`)
    await page.goto(`/#/flags/${flag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await expect(page.locator('[data-testid="segment-desc-input"]').first()).toBeVisible()
    await expect(page.locator('[data-testid="segment-rollout-input"]').first()).toBeVisible()
    await expect(page.locator('[data-testid="save-segment-btn"]').first()).toBeVisible()
    await expect(page.locator('[data-testid="delete-segment-btn"]').first()).toBeVisible()
    await deleteFlag(flag.id)
  })

  test('can update segment description and rollout', async ({ page }) => {
    const flag = await createFlag()
    await createSegment(flag.id, `init-${Date.now()}`)
    await page.goto(`/#/flags/${flag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.locator('input[data-testid="segment-desc-input"]').first().fill(`upd-${Date.now()}`)
    await page.locator('input[data-testid="segment-rollout-input"]').first().fill('75')
    await page.locator('[data-testid="save-segment-btn"]').first().click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    await deleteFlag(flag.id)
  })

  test('can delete a segment via UI', async ({ page }) => {
    const flag = await createFlag()
    await createSegment(flag.id, `del-${Date.now()}`)
    await page.goto(`/#/flags/${flag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.locator('[data-testid="delete-segment-btn"]').first().click()
    await expect(page.locator('.el-message-box')).toBeVisible({ timeout: 5000 })
    await page.locator('.el-message-box .el-button--primary').click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    await deleteFlag(flag.id)
  })

  test('existing constraints are displayed', async ({ page }) => {
    const flag = await createFlag()
    const seg = await createSegment(flag.id, `cstr-${Date.now()}`)
    await createConstraint(flag.id, seg.id)
    await page.goto(`/#/flags/${flag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.waitForSelector('input[data-testid="constraint-prop-input"]', { timeout: 5000 })
    await expect(page.locator('[data-testid="save-constraint-btn"]').first()).toBeVisible()
    await expect(page.locator('[data-testid="delete-constraint-btn"]').first()).toBeVisible()
    await deleteFlag(flag.id)
  })

  test('can create constraint via UI', async ({ page }) => {
    const flag = await createFlag()
    await createSegment(flag.id, `cstr-ui-${Date.now()}`)
    await page.goto(`/#/flags/${flag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.locator('input[data-testid="new-constraint-prop-input"]').first().fill('region')
    await page.locator('input[data-testid="new-constraint-value-input"]').first().fill('"EU"')
    await page.locator('[data-testid="add-constraint-btn"]').first().click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    await deleteFlag(flag.id)
  })

  test('can update constraint property', async ({ page }) => {
    const flag = await createFlag()
    const seg = await createSegment(flag.id, `cstr-upd-${Date.now()}`)
    await createConstraint(flag.id, seg.id)
    await page.goto(`/#/flags/${flag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.waitForSelector('input[data-testid="constraint-prop-input"]', { timeout: 5000 })
    await page.locator('input[data-testid="constraint-prop-input"]').first().fill('region')
    await page.locator('[data-testid="save-constraint-btn"]').first().click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    await deleteFlag(flag.id)
  })

  test('can delete constraint via UI', async ({ page }) => {
    const flag = await createFlag()
    const seg = await createSegment(flag.id, `cstr-del-${Date.now()}`)
    await createConstraint(flag.id, seg.id)
    await page.goto(`/#/flags/${flag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.waitForSelector('[data-testid="delete-constraint-btn"]', { timeout: 5000 })
    await page.locator('[data-testid="delete-constraint-btn"]').first().click()
    await expect(page.locator('.el-message-box')).toBeVisible({ timeout: 5000 })
    await page.locator('.el-message-box .el-button--primary').click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    await deleteFlag(flag.id)
  })

  test('history tab shows snapshots after changes', async ({ page }) => {
    const flag = await createFlagWithVariants()
    // Make a change to generate a snapshot
    await fetch(`http://localhost:18000/api/v1/flags/${flag.id}`, {
      method: 'PUT', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        description: `hist-${Date.now()}`,
        key: flag.key || '',
        dataRecordsEnabled: false,
        entityType: '',
        notes: ''
      })
    })
    await page.goto(`/#/flags/${flag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    // Wait for snapshot to be created
    await page.waitForTimeout(500)
    await page.locator('.el-tabs__item').filter({ hasText: 'History' }).click()
    await expect(page.locator('.snapshot-container').first()).toBeVisible({ timeout: 3000 })
    await deleteFlag(flag.id)
  })
})
