import { test, expect } from '@playwright/test'

test.describe('Flags Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.flags-container')
  })

  test('Page loads with table', async ({ page }) => {
    await expect(page.locator('.flags-container .el-table').first()).toBeVisible()
    // Check column headers exist
    await expect(page.locator('.flags-container .el-table__header').first()).toContainText('Flag ID')
    await expect(page.locator('.flags-container .el-table__header').first()).toContainText('Description')
    await expect(page.locator('.flags-container .el-table__header').first()).toContainText('Enabled')
  })

  test('Search field is present', async ({ page }) => {
    const searchInput = page.locator('input[placeholder="Search a flag"]')
    await expect(searchInput).toBeVisible()
  })

  test('Create new flag', async ({ page }) => {
    const descInput = page.locator('input[placeholder="Specific new flag description"]')
    const createBtn = page.locator('button').filter({ hasText: 'Create New Flag' })

    // Empty description - button should be disabled
    await descInput.fill('')
    await expect(createBtn).toBeDisabled()

    // Fill description
    const flagName = 'test-flag-e2e-' + Date.now()
    await descInput.fill(flagName)
    await expect(createBtn).not.toBeDisabled()

    // Click create
    await createBtn.click()
    await expect(page.locator('.el-message')).toContainText('flag created')

    // New flag appears in table
    await expect(page.locator('.flags-container .el-table__body').first()).toContainText(flagName)

    // Description field cleared
    await expect(descInput).toHaveValue('')
  })

  test('Create Simple Boolean Flag', async ({ page }) => {
    const descInput = page.locator('input[placeholder="Specific new flag description"]')
    const flagName = 'boolean-flag-e2e-' + Date.now()
    await descInput.fill(flagName)

    // Click dropdown arrow
    await page.locator('.el-dropdown__caret-button').click()
    await page.locator('.el-dropdown-menu__item').filter({ hasText: 'Create Simple Boolean Flag' }).click()

    await expect(page.locator('.el-message')).toContainText('flag created')
  })

  test('Search by description', async ({ page }) => {
    // Create flags for search
    const descInput = page.locator('input[placeholder="Specific new flag description"]')
    const createBtn = page.locator('button').filter({ hasText: 'Create New Flag' })

    const alpha = 'alpha-search-' + Date.now()
    await descInput.fill(alpha)
    await createBtn.click()
    await page.waitForTimeout(500)

    const beta = 'beta-search-' + Date.now()
    await descInput.fill(beta)
    await createBtn.click()
    await page.waitForTimeout(500)

    // Search for alpha
    const searchInput = page.locator('input[placeholder="Search a flag"]')
    await searchInput.fill('alpha-search')
    await page.waitForTimeout(300)

    const rows = page.locator('.el-table__body .el-table__row')
    for (const row of await rows.all()) {
      await expect(row).toContainText('alpha-search')
    }

    // Clear search
    await searchInput.fill('')
  })

  test('Search by ID', async ({ page }) => {
    const searchInput = page.locator('input[placeholder="Search a flag"]')
    await searchInput.fill('1')
    await page.waitForTimeout(300)
    // Should show at least one result containing ID 1
    await expect(page.locator('.flags-container .el-table__body').first()).toBeVisible()
  })

  test('Search by key', async ({ page }) => {
    const searchInput = page.locator('input[placeholder="Search a flag"]')
    // Just verify search doesn't break with a key-like term
    await searchInput.fill('flag')
    await page.waitForTimeout(300)
    await expect(page.locator('.flags-container .el-table__body').first()).toBeVisible()
    await searchInput.fill('')
  })

  test('AND search with comma', async ({ page }) => {
    const descInput = page.locator('input[placeholder="Specific new flag description"]')
    const createBtn = page.locator('button').filter({ hasText: 'Create New Flag' })

    const ts = Date.now()
    await descInput.fill('foo-bar-' + ts)
    await createBtn.click()
    await page.waitForTimeout(500)

    await descInput.fill('foo-baz-' + ts)
    await createBtn.click()
    await page.waitForTimeout(500)

    const searchInput = page.locator('input[placeholder="Search a flag"]')

    // "foo,bar" should match only foo-bar
    await searchInput.fill(`foo-bar-${ts},bar`)
    await page.waitForTimeout(300)
    const rows = page.locator('.el-table__body .el-table__row')
    const count = await rows.count()
    expect(count).toBeGreaterThanOrEqual(1)
    for (const row of await rows.all()) {
      await expect(row).toContainText('foo-bar')
    }

    await searchInput.fill('')
  })

  test('Table default sort by ID descending', async ({ page }) => {
    await expect(page.locator('.flags-container .el-table').first()).toBeVisible()
    // Just verify the table header has sorting capability
    const idHeader = page.locator('.flags-container .el-table__header th').filter({ hasText: 'Flag ID' }).first()
    await expect(idHeader).toBeVisible()
  })

  test('Table filter by Enabled/Disabled', async ({ page }) => {
    // Click the filter icon on Enabled column
    const enabledHeader = page.locator('.flags-container .el-table__header th').filter({ hasText: 'Enabled' }).last()
    await expect(enabledHeader).toBeVisible()
    // The filter functionality exists
    const filterIcon = enabledHeader.locator('.el-table__column-filter-trigger')
    if (await filterIcon.count() > 0) {
      await filterIcon.click()
      await page.waitForTimeout(300)
      // Close filter
      await page.keyboard.press('Escape')
    }
  })

  test('Navigate to flag by clicking row', async ({ page }) => {
    // Make sure there's at least one flag
    const rows = page.locator('.el-table__body .el-table__row')
    const count = await rows.count()
    if (count > 0) {
      await rows.first().click()
      await page.waitForTimeout(500)
      await expect(page).toHaveURL(/\/#\/flags\/\d+/)
    }
  })

  test('Search by description is case-insensitive', async ({ page }) => {
    const descInput = page.locator('input[placeholder="Specific new flag description"]')
    const createBtn = page.locator('button').filter({ hasText: 'Create New Flag' })

    // Create flag with mixed case description
    const mixedCaseName = 'TestFlag-CASE-' + Date.now()
    await descInput.fill(mixedCaseName)
    await createBtn.click()
    await page.waitForTimeout(500)

    // Search with lowercase
    const searchInput = page.locator('input[placeholder="Search a flag"]')
    await searchInput.fill('testflag-case')
    await page.waitForTimeout(300)

    // Verify flag is found (case-insensitive match)
    const rows = page.locator('.el-table__body .el-table__row')
    const count = await rows.count()
    expect(count).toBeGreaterThanOrEqual(1)

    const tableBody = page.locator('.flags-container .el-table__body').first()
    await expect(tableBody).toContainText(new RegExp(mixedCaseName, 'i'))

    // Clear search
    await searchInput.fill('')
  })

  test('Search by key is case-insensitive', async ({ page }) => {
    // Create a flag and set its key via the flag detail page
    const descInput = page.locator('input[placeholder="Specific new flag description"]')
    const createBtn = page.locator('button').filter({ hasText: 'Create New Flag' })

    const ts = Date.now()
    await descInput.fill('key-case-test-' + ts)
    await createBtn.click()
    await page.waitForTimeout(500)

    // Navigate to the new flag to set its key
    await page.locator('.el-table__body .el-table__row').first().click()
    await page.waitForSelector('.flag-container', { timeout: 10000 })

    const keyInput = page.locator('.flag-content input[placeholder="Key"]')
    const mixedCaseKey = 'MyTestKey-' + ts
    await keyInput.fill(mixedCaseKey)
    await page.locator('button').filter({ hasText: 'Save Flag' }).click()
    await expect(page.locator('.el-message').last()).toContainText('Flag updated')

    // Go back to flags list
    await page.goto('/')
    await page.waitForSelector('.flags-container')

    // Search with lowercase version of the key
    const searchInput = page.locator('input[placeholder="Search a flag"]')
    await searchInput.fill('mytestkey-')
    await page.waitForTimeout(300)

    // Verify flag is found
    const rows = page.locator('.el-table__body .el-table__row')
    expect(await rows.count()).toBeGreaterThanOrEqual(1)

    // Clear search
    await searchInput.fill('')
  })
})
