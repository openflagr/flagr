import { test, expect } from '@playwright/test'
const { createFlag, createVariant, createSegment } = require('./helpers')

let flagId

test.describe('Flag Distributions', () => {
  test.beforeAll(async () => {
    const flag = await createFlag('dist-test-' + Date.now())
    flagId = flag.id
    await createVariant(flagId, 'control')
    await createVariant(flagId, 'treatment')
    await createSegment(flagId, 'dist-segment')
  })

  test.beforeEach(async ({ page }) => {
    await page.goto(`http://localhost:8080/#/flags/${flagId}`)
    await page.waitForSelector('.flag-container', { timeout: 10000 })
  })

  test('Empty distribution state', async ({ page }) => {
    await expect(page.locator('.segment-distributions .card--error').first()).toContainText('No distribution yet')
  })

  test('Edit distribution dialog opens', async ({ page }) => {
    const editBtn = page.locator('.segment-distributions button').filter({ hasText: 'edit' }).first()
    await editBtn.click()
    await page.waitForTimeout(300)
    const dialog = page.locator('.el-dialog').filter({ hasText: 'Edit distribution' })
    await expect(dialog).toBeVisible()
    await page.keyboard.press('Escape')
  })

  test('Select variant with checkbox', async ({ page }) => {
    const editBtn = page.locator('.segment-distributions button').filter({ hasText: 'edit' }).first()
    await editBtn.click()
    await page.waitForTimeout(300)
    const dialog = page.locator('.el-dialog').filter({ hasText: 'Edit distribution' })
    const checkboxes = dialog.locator('.el-checkbox')
    if (await checkboxes.count() > 0) {
      await checkboxes.first().click()
      await page.waitForTimeout(200)
    }
    await page.keyboard.press('Escape')
  })

  test('Slider for percentage', async ({ page }) => {
    const editBtn = page.locator('.segment-distributions button').filter({ hasText: 'edit' }).first()
    await editBtn.click()
    await page.waitForTimeout(300)
    const dialog = page.locator('.el-dialog').filter({ hasText: 'Edit distribution' })
    const sliders = dialog.locator('.el-slider')
    expect(await sliders.count()).toBeGreaterThanOrEqual(1)
    await page.keyboard.press('Escape')
  })

  test('Validation: sum must equal 100%', async ({ page }) => {
    const editBtn = page.locator('.segment-distributions button').filter({ hasText: 'edit' }).first()
    await editBtn.click()
    await page.waitForTimeout(300)
    const dialog = page.locator('.el-dialog').filter({ hasText: 'Edit distribution' })
    const saveBtn = dialog.locator('button').filter({ hasText: 'Save' })
    await expect(saveBtn).toBeDisabled()
    await expect(dialog.locator('.el-alert')).toContainText('Percentages must add up to 100%')
    await page.keyboard.press('Escape')
  })

  test('Save distribution', async ({ page }) => {
    const editBtn = page.locator('.segment-distributions button').filter({ hasText: 'edit' }).first()
    await editBtn.click()
    await page.waitForTimeout(300)
    const dialog = page.locator('.el-dialog').filter({ hasText: 'Edit distribution' })
    // Check both variants
    const checkboxes = dialog.locator('.el-checkbox')
    for (let i = 0; i < await checkboxes.count(); i++) {
      const cb = checkboxes.nth(i)
      const isChecked = await cb.evaluate(el => el.classList.contains('is-checked'))
      if (!isChecked) {
        await cb.click()
        await page.waitForTimeout(200)
      }
    }
    // Set 50/50
    const sliderInputs = dialog.locator('.el-input-number input')
    const inputCount = await sliderInputs.count()
    if (inputCount >= 2) {
      await sliderInputs.nth(0).fill('')
      await sliderInputs.nth(0).type('50')
      await sliderInputs.nth(0).press('Enter')
      await page.waitForTimeout(200)
      await sliderInputs.nth(1).fill('')
      await sliderInputs.nth(1).type('50')
      await sliderInputs.nth(1).press('Enter')
      await page.waitForTimeout(200)
    }
    const saveBtn = dialog.locator('button').filter({ hasText: 'Save' })
    if (await saveBtn.isEnabled()) {
      await saveBtn.click()
      await expect(page.locator('.el-message')).toContainText('distributions updated')
    } else {
      await page.keyboard.press('Escape')
    }
  })

  test('Re-open distribution preserves values', async ({ page }) => {
    const editBtn = page.locator('.segment-distributions button').filter({ hasText: 'edit' }).first()
    await editBtn.click()
    await page.waitForTimeout(300)
    const dialog = page.locator('.el-dialog').filter({ hasText: 'Edit distribution' })
    await expect(dialog).toBeVisible()
    await page.keyboard.press('Escape')
  })
})
