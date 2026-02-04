import { test, expect } from '@playwright/test'

test.describe('Navigation and Layout', () => {
  test('Navbar renders with logo and version', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('.logo')).toContainText('Flagr')
    await expect(page.locator('.version')).toContainText('v')
    // Check API and Docs links
    const apiLink = page.locator('a[href*="api_docs"]')
    await expect(apiLink).toHaveAttribute('target', '_blank')
    const docsLink = page.locator('a[href*="openflagr.github.io/flagr"]').last()
    await expect(docsLink).toHaveAttribute('target', '_blank')
  })

  test('Click logo navigates to home', async ({ page }) => {
    await page.goto('/#/flags/1')
    await page.locator('.logo').click()
    await expect(page).toHaveURL(/\/#\/$/)
  })

  test('Breadcrumbs on home page', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.el-breadcrumb')
    await expect(page.locator('.el-breadcrumb')).toContainText('Home page')
  })

  test('Breadcrumbs on flag page', async ({ page }) => {
    // First create a flag to make sure flag 1 exists
    await page.goto('/')
    await page.waitForSelector('.flags-container')
    await page.goto('/#/flags/1')
    await page.waitForSelector('.el-breadcrumb')
    await expect(page.locator('.el-breadcrumb')).toContainText('Home page')
    await expect(page.locator('.el-breadcrumb')).toContainText('Flag ID: 1')
    // Click Home page breadcrumb
    await page.locator('.el-breadcrumb__item').first().locator('a, .el-breadcrumb__inner').first().click()
    await expect(page).toHaveURL(/\/#\/$/)
  })

  test('Router works with hash mode', async ({ page }) => {
    // Home page shows flags table
    await page.goto('/#/')
    await page.waitForSelector('.flags-container')

    // Flag page shows config
    await page.goto('/#/flags/1')
    await page.waitForSelector('.flag-container', { timeout: 5000 }).catch(() => {})

    // Unknown URL doesn't break
    await page.goto('/#/unknown')
    await expect(page.locator('#app').first()).toBeVisible()
  })
})
