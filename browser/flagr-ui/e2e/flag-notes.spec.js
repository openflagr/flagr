import { test, expect } from '@playwright/test'
const { createFlag } = require('./helpers')

let flagId

test.describe('Flag Notes', () => {
  test.beforeAll(async () => {
    const flag = await createFlag('notes-test-' + Date.now())
    flagId = flag.id
  })

  test.beforeEach(async ({ page }) => {
    await page.goto(`http://localhost:8080/#/flags/${flagId}`)
    await page.waitForSelector('.flag-container', { timeout: 10000 })
  })

  test('Edit/view button toggles mode', async ({ page }) => {
    const toggleBtn = page.locator('button').filter({ hasText: /edit|view/ }).first()
    await expect(toggleBtn).toBeVisible()
    await toggleBtn.click()
    await expect(toggleBtn).toContainText('view')
    await toggleBtn.click()
    await expect(toggleBtn).toContainText('edit')
  })

  test('Editor hidden when no notes and editor closed', async ({ page }) => {
    const editor = page.locator('#editor')
    await expect(editor).not.toBeVisible()
  })

  test('Markdown editor textarea appears in edit mode', async ({ page }) => {
    const toggleBtn = page.locator('button').filter({ hasText: 'edit' }).first()
    await toggleBtn.click()
    await page.waitForTimeout(300)
    const textarea = page.locator('#editor textarea')
    await expect(textarea).toBeVisible()
  })

  test('Markdown preview renders', async ({ page }) => {
    const toggleBtn = page.locator('button').filter({ hasText: 'edit' }).first()
    await toggleBtn.click()
    await page.waitForTimeout(300)
    const textarea = page.locator('#editor textarea')
    await textarea.fill('**bold text**')
    await textarea.blur()
    await page.waitForTimeout(300)
    const preview = page.locator('.markdown-body')
    await expect(preview).toBeVisible()
  })

  test('XSS filtering', async ({ page }) => {
    const toggleBtn = page.locator('button').filter({ hasText: 'edit' }).first()
    await toggleBtn.click()
    await page.waitForTimeout(300)
    const textarea = page.locator('#editor textarea')
    await textarea.fill('<script>alert(1)</script>')
    await textarea.blur()
    await page.waitForTimeout(300)
    const scriptTag = page.locator('.markdown-body script')
    expect(await scriptTag.count()).toBe(0)
  })

  test('Save notes via Save Flag', async ({ page }) => {
    const toggleBtn = page.locator('button').filter({ hasText: 'edit' }).first()
    await toggleBtn.click()
    await page.waitForTimeout(300)
    const textarea = page.locator('#editor textarea')
    const noteText = 'Test note ' + Date.now()
    await textarea.fill(noteText)
    await textarea.blur()
    await page.waitForTimeout(300)
    await page.locator('button').filter({ hasText: 'Save Flag' }).click()
    await expect(page.locator('.el-message')).toContainText('Flag updated')
    await page.reload()
    await page.waitForSelector('.flag-container')
    await page.waitForTimeout(500)
    const preview = page.locator('.markdown-body')
    await expect(preview).toBeVisible({ timeout: 5000 })
    await expect(preview).toContainText(noteText)
  })
})
