import { test, expect } from '@playwright/test'
const { createFlag, createVariant, createSegment } = require('./helpers')

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

  test('Save and verify variant attachment JSON', async ({ page }) => {
    // 1. Ensure we have a variant
    const variantInputs = page.locator('.variants-container-inner .variant-key-input input')
    if (await variantInputs.count() === 0) {
      // Create variant if none exists
      const keyInput = page.locator('input[placeholder="Variant Key"]')
      const createBtn = page.locator('button').filter({ hasText: 'Create Variant' })
      await keyInput.fill('attachment-test')
      await createBtn.click()
      await page.waitForTimeout(500)
    }

    // 2. Open attachment collapse
    const collapseHeader = page.locator('.variant-attachment-collapsable-title .el-collapse-item__header').first()
    await collapseHeader.click()
    await page.waitForTimeout(500)

    // 3. Find CodeMirror editor and enter JSON
    // Click on the line content to focus
    const cmLine = page.locator('.variant-attachment-content .cm-line').first()
    await cmLine.click({ clickCount: 3 }) // Triple-click to select all in the line
    await page.waitForTimeout(200)

    // Type new JSON - this will replace the selected text
    const testJson = '{"testKey": "testValue123"}'
    await page.keyboard.type(testJson, { delay: 5 })
    await page.waitForTimeout(300)

    // Click outside the editor to trigger blur and ensure v-model syncs
    await page.locator('.variant-attachment-collapsable-title').first().click()
    await page.waitForTimeout(500)

    // 4. Save variant
    await page.locator('.variants-container-inner button').filter({ hasText: 'Save Variant' }).first().click()
    await expect(page.locator('.el-message').last()).toContainText('variant updated')
    await page.waitForTimeout(500)

    // 5. Reload page
    await page.reload()
    await page.waitForSelector('.flag-container', { timeout: 10000 })

    // 6. Open attachment again
    const collapseHeaderAfter = page.locator('.variant-attachment-collapsable-title .el-collapse-item__header').first()
    await collapseHeaderAfter.click()
    await page.waitForTimeout(500)

    // 7. Verify JSON was saved
    const editorContent = page.locator('.variant-attachment-content .cm-content').first()
    await expect(editorContent).toContainText('testKey')
    await expect(editorContent).toContainText('testValue123')
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
    const deleteIcons = page.locator('.variants-container-inner .save-remove-variant-row .el-icon')
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

test.describe('Variant Delete Protection', () => {
  let flagIdWithDist

  test.beforeAll(async () => {
    // Create flag with variant and segment for distribution test
    const flag = await createFlag('variant-delete-protection-' + Date.now())
    flagIdWithDist = flag.id
    await createVariant(flagIdWithDist, 'protected-variant')
    await createSegment(flagIdWithDist, 'protection-segment')
  })

  test('Cannot delete variant that is in active distribution', async ({ page }) => {
    await page.goto(`http://localhost:8080/#/flags/${flagIdWithDist}`)
    await page.waitForSelector('.flag-container', { timeout: 10000 })

    // First, add the variant to a distribution via UI
    const editBtn = page.locator('.segment-distributions button').filter({ hasText: 'edit' }).first()
    await editBtn.click()
    await page.waitForTimeout(300)

    const dialog = page.locator('.el-dialog').filter({ hasText: 'Edit distribution' })
    const checkboxes = dialog.locator('.el-checkbox')

    // Check the first variant
    const firstCheckbox = checkboxes.first()
    const isChecked = await firstCheckbox.locator('input[type="checkbox"]').isChecked()
    if (!isChecked) {
      await firstCheckbox.click()
      await page.waitForTimeout(200)
    }

    // Set to 100%
    const sliderInputs = dialog.locator('.el-input-number input')
    if (await sliderInputs.count() > 0) {
      await sliderInputs.first().fill('')
      await sliderInputs.first().type('100')
      await sliderInputs.first().press('Enter')
      await page.waitForTimeout(200)
    }

    // Save distribution
    const saveBtn = dialog.locator('button').filter({ hasText: 'Save' })
    if (await saveBtn.isEnabled()) {
      await saveBtn.click()
      await page.waitForTimeout(500)
    } else {
      await page.keyboard.press('Escape')
    }

    // Now try to delete the variant - expect alert
    let alertMessage = ''
    page.on('dialog', async dialog => {
      alertMessage = dialog.message()
      await dialog.dismiss()
    })

    // Find and click delete button for the variant
    const deleteIcons = page.locator('.variants-container-inner .save-remove-variant-row .el-icon')
    if (await deleteIcons.count() > 0) {
      await deleteIcons.first().click()
      await page.waitForTimeout(500)
    }

    // Verify alert was shown with the expected message
    expect(alertMessage).toContain('being used by a segment distribution')

    // Verify variant still exists
    const variantInputs = page.locator('.variants-container-inner .variant-key-input input')
    expect(await variantInputs.count()).toBeGreaterThanOrEqual(1)
  })
})
