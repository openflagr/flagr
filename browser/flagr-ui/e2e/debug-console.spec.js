import { test, expect } from '@playwright/test'
const { createFlag } = require('./helpers')

let flagId

test.describe('Debug Console', () => {
  test.beforeAll(async () => {
    const flag = await createFlag('debug-test-' + Date.now())
    flagId = flag.id
  })

  test.beforeEach(async ({ page }) => {
    await page.goto(`http://localhost:8080/#/flags/${flagId}`)
    await page.waitForSelector('.flag-container', { timeout: 10000 })
  })

  test('Debug Console card renders', async ({ page }) => {
    await expect(page.locator('.dc-container')).toBeVisible()
    await expect(page.locator('.dc-container')).toContainText('Debug Console')
  })

  test('Evaluation collapse', async ({ page }) => {
    const evalItem = page.locator('.dc-container .el-collapse-item').first()
    await evalItem.locator('.el-collapse-item__header').click()
    await page.waitForTimeout(300)
    await expect(evalItem).toContainText('Request')
    await expect(evalItem).toContainText('Response')
    await expect(evalItem.locator('button').filter({ hasText: /^POST \/api\/v1\/evaluation$/ })).toBeVisible()
  })

  test('POST evaluation', async ({ page }) => {
    const evalItem = page.locator('.dc-container .el-collapse-item').first()
    await evalItem.locator('.el-collapse-item__header').click()
    await page.waitForTimeout(300)
    await evalItem.locator('button').filter({ hasText: /^POST \/api\/v1\/evaluation$/ }).click()
    await page.waitForTimeout(1000)
    const msg = page.locator('.el-message')
    await expect(msg).toContainText(/evaluation (success|error)/)
  })

  test('Batch Evaluation collapse', async ({ page }) => {
    const batchItem = page.locator('.dc-container .el-collapse-item').nth(1)
    await batchItem.locator('.el-collapse-item__header').click()
    await page.waitForTimeout(300)
    await expect(page.locator('button').filter({ hasText: 'POST /api/v1/evaluation/batch' })).toBeVisible()
  })

  test('POST batch evaluation', async ({ page }) => {
    const batchItem = page.locator('.dc-container .el-collapse-item').nth(1)
    await batchItem.locator('.el-collapse-item__header').click()
    await page.waitForTimeout(300)
    await page.locator('button').filter({ hasText: 'POST /api/v1/evaluation/batch' }).click()
    await page.waitForTimeout(1000)
    const msg = page.locator('.el-message')
    await expect(msg).toContainText(/evaluation (success|error)/)
  })
})
