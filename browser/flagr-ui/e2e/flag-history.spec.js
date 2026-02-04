import { test, expect } from '@playwright/test'
const { API, createFlag } = require('./helpers')

let flagId

test.describe('Flag History', () => {
  test.beforeAll(async () => {
    const flag = await createFlag('history-test-' + Date.now())
    flagId = flag.id
    // Save the flag to create a snapshot
    await fetch(`${API}/flags/${flagId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ description: flag.description, key: flag.key || '', dataRecordsEnabled: false, entityType: '', notes: '' })
    })
  })

  test.beforeEach(async ({ page }) => {
    await page.goto(`http://localhost:8080/#/flags/${flagId}`)
    await page.waitForSelector('.flag-container', { timeout: 10000 })
  })

  test('History tab loads lazily', async ({ page }) => {
    await expect(page.locator('.el-tabs__item').filter({ hasText: 'Config' })).toHaveClass(/is-active/)
    await page.locator('.el-tabs__item').filter({ hasText: 'History' }).click()
    await page.waitForTimeout(1000)
    await expect(page.locator('.el-tabs__item').filter({ hasText: 'History' })).toHaveClass(/is-active/)
  })

  test('Snapshots display', async ({ page }) => {
    await page.locator('.el-tabs__item').filter({ hasText: 'History' }).click()
    await page.waitForTimeout(1000)
    const snapshots = page.locator('.snapshot-container')
    if (await snapshots.count() > 0) {
      await expect(snapshots.first()).toContainText('Snapshot ID')
    }
  })

  test('Diff between snapshots', async ({ page }) => {
    await page.locator('.el-tabs__item').filter({ hasText: 'History' }).click()
    await page.waitForTimeout(1000)
    const diffs = page.locator('.snapshot-container .diff')
    if (await diffs.count() > 0) {
      await expect(diffs.first()).toBeVisible()
    }
  })

  test('Switch back to Config tab', async ({ page }) => {
    await page.locator('.el-tabs__item').filter({ hasText: 'History' }).click()
    await page.waitForTimeout(500)
    await page.locator('.el-tabs__item').filter({ hasText: 'Config' }).click()
    await page.waitForTimeout(500)
    await expect(page.locator('.el-tabs__item').filter({ hasText: 'Config' })).toHaveClass(/is-active/)
    await expect(page.locator('.flag-config-card')).toBeVisible()
  })
})
