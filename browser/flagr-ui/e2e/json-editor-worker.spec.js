import { test, expect } from '@playwright/test'
const { createFlag } = require('./helpers')

/**
 * Test for json-editor-vue (CodeMirror-based editor).
 *
 * After replacing vue3-json-editor with json-editor-vue,
 * there should be no Web Worker warnings. The new library
 * uses CodeMirror instead of ACE editor and doesn't have
 * the Blob/Worker issues with Webpack 5.
 */

let flagId

test.describe('JSON Editor (json-editor-vue)', () => {
  test.beforeAll(async () => {
    const flag = await createFlag('json-editor-test-' + Date.now())
    flagId = flag.id
  })

  test('no worker loading errors', async ({ page }) => {
    const workerWarnings = []

    // Capture console warnings
    page.on('console', msg => {
      if (msg.type() === 'warning' && msg.text().includes('Could not load worker')) {
        workerWarnings.push(msg.text())
      }
    })

    await page.goto(`http://localhost:8080/#/flags/${flagId}`)
    await page.waitForSelector('.flag-container', { timeout: 10000 })

    // Open Evaluation collapse to trigger json-editor rendering
    const evalItem = page.locator('.dc-container .el-collapse-item').first()
    await evalItem.locator('.el-collapse-item__header').click()
    await page.waitForTimeout(2000)

    console.log(`Worker warnings count: ${workerWarnings.length}`)

    // After replacing with json-editor-vue, there should be no worker warnings
    expect(workerWarnings.length,
      'Worker warnings found - json-editor-vue should not have worker issues'
    ).toBe(0)
  })

  test('json editor renders correctly', async ({ page }) => {
    await page.goto(`http://localhost:8080/#/flags/${flagId}`)
    await page.waitForSelector('.flag-container', { timeout: 10000 })

    const evalItem = page.locator('.dc-container .el-collapse-item').first()
    await evalItem.locator('.el-collapse-item__header').click()
    await page.waitForTimeout(1000)

    // json-editor-vue uses .jse-main and CodeMirror (.cm-content)
    await expect(page.locator('.jse-main').first()).toBeVisible()
    await expect(page.locator('.cm-content').first()).toBeVisible()

    // Verify JSON content is displayed
    const editorText = await page.locator('.cm-content').first().textContent()
    expect(editorText).toContain('entityID')
  })
})
