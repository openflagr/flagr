import { test, expect } from '@playwright/test'

test.describe('Full E2E Workflow', () => {
  test('Complete flag lifecycle', async ({ page }) => {
    page.on('dialog', dialog => dialog.accept())

    // 1. Go to home, create flag
    await page.goto('http://localhost:8080/')
    await page.waitForSelector('.flags-container')

    const descInput = page.locator('input[placeholder="Specific new flag description"]')
    const createBtn = page.locator('button').filter({ hasText: 'Create New Flag' })
    const flagName = 'e2e-workflow-' + Date.now()

    await descInput.fill(flagName)
    await createBtn.click()
    await expect(page.locator('.el-message')).toContainText('flag created')
    await page.waitForTimeout(1000)

    // 2. Navigate to flag
    await page.locator('.flags-container .el-table__body .el-table__row').first().click()
    await page.waitForSelector('.flag-container', { timeout: 10000 })
    await expect(page.locator('.flag-config-card')).toBeVisible()

    // 3. Enable flag
    const switchEl = page.locator('.flag-config-card .el-card-header .el-switch')
    await switchEl.click()
    await page.waitForTimeout(1000)

    // 4-5. Create variants
    const variantInput = page.locator('input[placeholder="Variant Key"]')
    const createVarBtn = page.locator('button').filter({ hasText: 'Create Variant' })

    await variantInput.fill('control')
    await createVarBtn.click()
    await expect(page.locator('.el-message').last()).toContainText('new variant created')
    await page.waitForTimeout(1000)

    await variantInput.fill('treatment')
    await createVarBtn.click()
    await expect(page.locator('.el-message').last()).toContainText('new variant created')
    await page.waitForTimeout(1000)

    // 6. Create segment
    await page.locator('button').filter({ hasText: 'New Segment' }).click()
    const segDialog = page.locator('.el-dialog').filter({ hasText: 'Create segment' })
    await segDialog.locator('input[placeholder="Segment description"]').fill('all-users')
    await segDialog.locator('button').filter({ hasText: 'Create Segment' }).click()
    await expect(page.locator('.el-message').last()).toContainText('new segment created')
    await page.waitForTimeout(1000)

    // 7. Add constraint (values must be quoted for the backend parser)
    const segment = page.locator('.segment').first()
    const propInput = segment.locator('.constraints input[placeholder="Property"]').last()
    await propInput.fill('env')
    const constraintInputs = segment.locator('.constraints .el-col .el-input input')
    const lastConstraintInput = constraintInputs.last()
    await lastConstraintInput.fill('"production"')
    await segment.locator('button').filter({ hasText: 'Add Constraint' }).click()
    await expect(page.locator('.el-message').last()).toContainText('new constraint created')
    await page.waitForTimeout(1000)

    // 9. Create tag
    await page.locator('button').filter({ hasText: '+ New Tag' }).click()
    await page.waitForTimeout(300)
    const tagInput = page.locator('.tag-key-input input')
    await tagInput.fill('experiment')
    await tagInput.press('Enter')
    await page.waitForTimeout(1000)

    // 11. Save Flag
    await page.locator('button').filter({ hasText: 'Save Flag' }).click()
    await expect(page.locator('.el-message').last()).toContainText('Flag updated')
    await page.waitForTimeout(1000)

    // 12. Debug Console - POST evaluation (scope to first collapse item to avoid matching batch button)
    const evalCollapse = page.locator('.dc-container .el-collapse-item').first()
    await evalCollapse.locator('.el-collapse-item__header').click()
    await page.waitForTimeout(300)
    await evalCollapse.locator('button').filter({ hasText: /^\s*POST \/api\/v1\/evaluation\s*$/ }).click()
    await page.waitForTimeout(1000)

    // 14. History tab
    await page.locator('.el-tabs__item').filter({ hasText: 'History' }).click()
    await page.waitForTimeout(1000)

    // 15. Back to home
    await page.locator('.logo').click()
    await page.waitForSelector('.flags-container')
    await expect(page.locator('.flags-container .el-table__body').first()).toContainText(flagName)

    // 16. Search
    const searchInput = page.locator('input[placeholder="Search a flag"]')
    await searchInput.fill(flagName)
    await page.waitForTimeout(300)
    await expect(page.locator('.flags-container .el-table__body').first()).toContainText(flagName)
    await searchInput.fill('')
    await page.waitForTimeout(300)

    // 17. Delete flag
    await page.locator('.flags-container .el-table__body .el-table__row').filter({ hasText: flagName }).first().click()
    await page.waitForSelector('.flag-container', { timeout: 10000 })

    const deleteBtn = page.locator('button').filter({ hasText: 'Delete Flag' })
    await deleteBtn.click()
    const confirmBtn = page.locator('.el-dialog').locator('button').filter({ hasText: 'Confirm' })
    if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await confirmBtn.click()
    }
    await page.waitForTimeout(1000)
    await page.waitForSelector('.flags-container', { timeout: 5000 })
  })
})
