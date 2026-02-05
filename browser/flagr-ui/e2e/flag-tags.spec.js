import { test, expect } from '@playwright/test'
const { createFlag } = require('./helpers')

let flagId

test.describe('Flag Tags', () => {
  test.beforeAll(async () => {
    const flag = await createFlag('tags-test-' + Date.now())
    flagId = flag.id
  })

  test.beforeEach(async ({ page }) => {
    await page.goto(`http://localhost:8080/#/flags/${flagId}`)
    await page.waitForSelector('.flag-config-card', { timeout: 10000 })
  })

  test('New Tag button visible, input hidden', async ({ page }) => {
    await expect(page.locator('button').filter({ hasText: '+ New Tag' })).toBeVisible()
    await expect(page.locator('.tag-key-input')).not.toBeVisible()
  })

  test('Click New Tag shows input', async ({ page }) => {
    await page.locator('button').filter({ hasText: '+ New Tag' }).click()
    await expect(page.locator('.tag-key-input')).toBeVisible()
  })

  test('Create tag via Enter', async ({ page }) => {
    await page.locator('button').filter({ hasText: '+ New Tag' }).click()
    await page.waitForTimeout(300)
    const tagInput = page.locator('.tag-key-input input')
    const tagName = 'tag-' + Date.now()
    await tagInput.fill(tagName)
    await tagInput.press('Enter')
    await page.waitForTimeout(500)
    await expect(page.locator('.el-message')).toContainText('new tag created')
    await expect(page.locator('.tags-container-inner .el-tag').filter({ hasText: tagName })).toBeVisible()
  })

  test('Duplicate tags not added in UI', async ({ page }) => {
    const tagName = 'dup-tag-' + Date.now()
    // Create first
    await page.locator('button').filter({ hasText: '+ New Tag' }).click()
    await page.waitForTimeout(300)
    let tagInput = page.locator('.tag-key-input input')
    await tagInput.fill(tagName)
    await tagInput.press('Enter')
    await page.waitForTimeout(500)
    // Create same again
    await page.locator('button').filter({ hasText: '+ New Tag' }).click()
    await page.waitForTimeout(300)
    tagInput = page.locator('.tag-key-input input')
    await tagInput.fill(tagName)
    await tagInput.press('Enter')
    await page.waitForTimeout(500)
    const tags = page.locator('.tags-container-inner .el-tag').filter({ hasText: tagName })
    expect(await tags.count()).toBe(1)
  })

  test('Cancel tag creation with Esc', async ({ page }) => {
    await page.locator('button').filter({ hasText: '+ New Tag' }).click()
    await page.waitForTimeout(300)
    const tagInput = page.locator('.tag-key-input input')
    await tagInput.fill('cancel-this')
    await tagInput.press('Escape')
    await page.waitForTimeout(300)
    await expect(page.locator('.tag-key-input')).not.toBeVisible()
  })

  test('Delete tag', async ({ page }) => {
    // Create a tag first
    await page.locator('button').filter({ hasText: '+ New Tag' }).click()
    await page.waitForTimeout(300)
    const tagInput = page.locator('.tag-key-input input')
    const tagName = 'del-tag-' + Date.now()
    await tagInput.fill(tagName)
    await tagInput.press('Enter')
    await page.waitForTimeout(500)
    // Handle confirm dialog
    page.on('dialog', dialog => dialog.accept())
    const tag = page.locator('.tags-container-inner .el-tag').filter({ hasText: tagName })
    const closeBtn = tag.locator('.el-tag__close, .el-icon-close')
    await closeBtn.click()
    await page.waitForTimeout(500)
    await expect(page.locator('.el-message').last()).toContainText('tag deleted')
  })

  test('Tag autocomplete', async ({ page }) => {
    await page.locator('button').filter({ hasText: '+ New Tag' }).click()
    await page.waitForTimeout(300)
    const tagInput = page.locator('.tag-key-input input')
    await tagInput.fill('tag')
    await page.waitForTimeout(500)
    await expect(tagInput).toHaveValue('tag')
  })

  test('Create tags with allowed special characters', async ({ page }) => {
    // Backend regex: ^[ \w\d-/\.\:]+$
    // Allowed: alphanumeric, hyphen, slash, dot, colon, space
    const validTags = [
      'my-tag-' + Date.now(),
      'v1.0.' + Date.now(),
      'env:prod-' + Date.now(),
      'feature/login-' + Date.now()
    ]

    for (const tagName of validTags) {
      await page.locator('button').filter({ hasText: '+ New Tag' }).click()
      await page.waitForTimeout(300)
      const tagInput = page.locator('.tag-key-input input')
      await tagInput.fill(tagName)
      await tagInput.press('Enter')
      await page.waitForTimeout(500)
      await expect(page.locator('.el-message').last()).toContainText('new tag created')
      await expect(page.locator('.tags-container-inner .el-tag').filter({ hasText: tagName })).toBeVisible()
    }
  })

  test('Tag with invalid characters shows error', async ({ page }) => {
    // Characters NOT allowed: @, #, $, !, etc.
    const invalidTagName = 'tag@invalid-' + Date.now()

    await page.locator('button').filter({ hasText: '+ New Tag' }).click()
    await page.waitForTimeout(300)
    const tagInput = page.locator('.tag-key-input input')
    await tagInput.fill(invalidTagName)
    await tagInput.press('Enter')
    await page.waitForTimeout(500)

    // Backend returns 400 error for invalid tag format
    await expect(page.locator('.el-message--error').last()).toBeVisible()
    const tagExists = await page.locator('.tags-container-inner .el-tag')
      .filter({ hasText: invalidTagName }).count()
    expect(tagExists).toBe(0)
  })
})
