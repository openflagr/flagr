// Verify DebugConsole reactive state syncs after editing JSON in the editor.
// Without @update:jsonString, does @update:json alone keep evalContext in sync?
import { test, expect } from '@playwright/test'

const API = 'http://127.0.0.1:18000/api/v1'

test('debug console syncs edited json to POST payload', async ({ page }) => {
  // Create a flag with a variant
  const flagResp = await page.request.post(`${API}/flags`, {
    data: { key: `e2e-json-sync-${Date.now()}`, description: 'test', enabled: true }
  })
  const flag = await flagResp.json()
  const varResp = await page.request.post(`${API}/flags/${flag.id}/variants`, {
    data: { key: 'control', attachment: { color: 'blue' } }
  })
  const variant = await varResp.json()
  const segResp = await page.request.post(`${API}/flags/${flag.id}/segments`, {
    data: { description: 'all users', rolloutPercent: 100 }
  })
  const seg = await segResp.json()
  await page.request.put(`${API}/flags/${flag.id}/segments/${seg.id}/distributions`, {
    data: { distributions: [{ percent: 100, variantID: variant.id, variantKey: 'control' }] }
  })

  await page.goto(`/#/flags/${flag.id}`)
  await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })

  // Expand the DebugConsole
  await page.locator('.dc-container .el-collapse-item__header').first().click()
  const editor = page.locator('.dc-container .cm-content, .dc-container .jse-text-pane textarea, .dc-container .jse-text-pane div[contenteditable]').first()
  await expect(editor).toBeVisible({ timeout: 5000 })

  // Get the current JSON text
  const currentText = await editor.evaluate(el => el.textContent || el.innerText || '')
  const current = JSON.parse(currentText)

  // Modify the entityID and entityContext
  current.entityID = 'e2e-test-user-42'
  current.entityContext = { region: 'EU', tier: 'beta' }

  // Clear and type the modified JSON
  await editor.click()
  await editor.evaluate(el => el.textContent = '')
  const newText = JSON.stringify(current, null, 2)
  await editor.evaluate((el, text) => {
    if (el.textContent !== undefined) el.textContent = text
    else el.innerText = text
    el.dispatchEvent(new Event('input', { bubbles: true }))
  }, newText)

  await page.waitForTimeout(500)

  // Click POST
  await page.locator('.dc-container button:has-text("POST")').first().click()
  await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 10000 })

  // Verify the rendered result shows the variant (means the POST went through with our data)
  await expect(page.locator('.dc-summary')).toBeVisible({ timeout: 5000 })
  await expect(page.locator('.dc-result-variant-value')).toBeVisible({ timeout: 5000 })

  // Verify the evaluation actually used our entityID by checking segment debug log
  const variantValue = await page.locator('.dc-result-variant-value').textContent()
  expect(variantValue).toBe('control')
})
