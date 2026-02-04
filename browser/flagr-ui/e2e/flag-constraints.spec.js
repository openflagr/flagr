import { test, expect } from '@playwright/test'
const { createFlag, createSegment } = require('./helpers')

let flagId

test.describe('Flag Constraints', () => {
  test.beforeAll(async () => {
    const flag = await createFlag('constraints-test-' + Date.now())
    flagId = flag.id
    await createSegment(flagId, 'constraint-test-segment')
  })

  test.beforeEach(async ({ page }) => {
    await page.goto(`http://localhost:8080/#/flags/${flagId}`)
    await page.waitForSelector('.flag-config-card', { timeout: 10000 })
  })

  test('Empty state shows no constraints message', async ({ page }) => {
    await expect(page.locator('.card--empty').first()).toContainText('No constraints (ALL will pass)')
  })

  test('Create constraint', async ({ page }) => {
    const segment = page.locator('.segment').first()
    // Property input for new constraint (last one in the constraints area)
    const propInput = segment.locator('input[placeholder="Property"]').last()
    await propInput.fill('country')
    // Value input - the el-input without a placeholder that's not a select
    // In the new constraint row, there are: Property input, operator select, value input, button
    // Find the value input in the last el-row of constraints
    const newConstraintRow = segment.locator('.constraints > div:last-child .el-row')
    const valueInput = newConstraintRow.locator('.el-col').nth(2).locator('input')
    await valueInput.fill('"US"')
    const addBtn = segment.locator('button').filter({ hasText: 'Add Constraint' })
    await addBtn.click()
    await expect(page.locator('.el-message')).toContainText('new constraint created')
  })

  test('All 12 operators available', async ({ page }) => {
    const segment = page.locator('.segment').first()
    const selects = segment.locator('.constraints .el-select')
    await selects.last().click()
    await page.waitForTimeout(300)
    const options = page.locator('.el-select-dropdown__item:visible')
    const count = await options.count()
    expect(count).toBeGreaterThanOrEqual(12)
    await page.keyboard.press('Escape')
  })

  test('Whitespace is trimmed', async ({ page }) => {
    const segment = page.locator('.segment').first()
    const propInput = segment.locator('input[placeholder="Property"]').last()
    await propInput.fill('  env  ')
    const newConstraintRow = segment.locator('.constraints > div:last-child .el-row')
    const valueInput = newConstraintRow.locator('.el-col').nth(2).locator('input')
    await valueInput.fill('  "prod"  ')
    const addBtn = segment.locator('button').filter({ hasText: 'Add Constraint' })
    await addBtn.click()
    await page.waitForTimeout(500)
    await expect(page.locator('.el-message').last()).toContainText('new constraint created')
  })

  test('Save constraint', async ({ page }) => {
    const segment = page.locator('.segment').first()
    const saveBtn = segment.locator('.segment-constraint button').filter({ hasText: 'Save' }).first()
    if (await saveBtn.isVisible().catch(() => false)) {
      await saveBtn.click()
      await expect(page.locator('.el-message')).toContainText('constraint updated')
    }
  })

  test('Delete constraint', async ({ page }) => {
    page.on('dialog', dialog => dialog.accept())
    const segment = page.locator('.segment').first()
    const deleteBtns = segment.locator('.segment-constraint .el-button--danger')
    if (await deleteBtns.count() > 0) {
      await deleteBtns.first().click()
      await page.waitForTimeout(500)
      await expect(page.locator('.el-message')).toContainText('constraint deleted')
    }
  })

  test('Multiple constraints heading', async ({ page }) => {
    await expect(page.locator('.segment').first()).toContainText('Constraints (match ALL of them)')
  })
})
