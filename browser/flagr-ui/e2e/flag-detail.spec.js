import { test, expect } from '@playwright/test'

const API = 'http://localhost:18000/api/v1'

async function createFlag(opts = {}) {
  const r = await fetch(`${API}/flags`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ description: opts.description || `e2e-${Date.now()}` })
  })
  const flag = await r.json()
  await fetch(`${API}/flags/${flag.id}/variants`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ key: 'control' })
  })
  await fetch(`${API}/flags/${flag.id}/variants`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ key: 'test' })
  })
  return flag
}

async function deleteFlag(flagId) {
  try { await fetch(`${API}/flags/${flagId}`, { method: 'DELETE' }) } catch {}
}

async function createSegment(flagId, desc) {
  const r = await fetch(`${API}/flags/${flagId}/segments`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ description: desc, rolloutPercent: 100 })
  })
  return r.json()
}

async function createConstraint(flagId, segId) {
  const r = await fetch(`${API}/flags/${flagId}/segments/${segId}/constraints`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ property: 'country', operator: 'EQ', value: '"US"' })
  })
  return r.json()
}

function sleep(ms) { return new Promise(r => setTimeout(r, ms)) }

test.describe('Flag detail page', () => {
  let testFlag

  test.beforeAll(async () => {
    testFlag = await createFlag({ description: `e2e-shared-${Date.now()}` })
  })

  test.afterAll(async () => {
    await deleteFlag(testFlag.id)
  })

  test('loads and displays flag details', async ({ page }) => {
    await page.goto(`/#/flags/${testFlag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })
  })

  test('shows all sections', async ({ page }) => {
    await page.goto(`/#/flags/${testFlag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await expect(page.locator('h2:has-text("Variants")')).toBeVisible()
    await expect(page.locator('h2:has-text("Segments")').first()).toBeVisible()
    await expect(page.locator('h2:has-text("Flag Settings")')).toBeVisible()
  })

  test('can toggle flag enabled state', async ({ page }) => {
    await page.goto(`/#/flags/${testFlag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.locator('[data-testid="flag-enable-switch"]').click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
  })

  test('can update flag key and save', async ({ page }) => {
    await page.goto(`/#/flags/${testFlag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.locator('input[data-testid="flag-key-input"]').fill(`key-${Date.now()}`)
    await page.locator('[data-testid="save-flag-btn"]').click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
  })

  test('can update variant key and save', async ({ page }) => {
    await page.goto(`/#/flags/${testFlag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.locator('input[data-testid="variant-key-input"]').first().fill(`vk-${Date.now()}`)
    await page.locator('[data-testid="save-variant-btn"]').first().click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
  })

  test('can create a new variant', async ({ page }) => {
    await page.goto(`/#/flags/${testFlag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.locator('input[data-testid="new-variant-input"]').fill(`v-${Date.now()}`)
    await page.locator('[data-testid="create-variant-btn"]').click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
  })

  test('can delete a variant', async ({ page }) => {
    await fetch(`${API}/flags/${testFlag.id}/variants`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ key: `del-v-${Date.now()}` })
    })
    await page.goto(`/#/flags/${testFlag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.locator('[data-testid="delete-variant-btn"]').last().click()
    await expect(page.locator('.el-message-box')).toBeVisible({ timeout: 5000 })
    await page.locator('.el-message-box .el-button--primary').click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
  })

  test('segment cards display correctly', async ({ page }) => {
    const flag = await createFlag({ description: `seg-display-${Date.now()}` })
    await createSegment(flag.id, `display-${Date.now()}`)
    await page.goto(`/#/flags/${flag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await expect(page.locator('[data-testid="segment-desc-input"]').first()).toBeVisible()
    await expect(page.locator('[data-testid="segment-rollout-input"]').first()).toBeVisible()
    await expect(page.locator('[data-testid="save-segment-btn"]').first()).toBeVisible()
    await expect(page.locator('[data-testid="delete-segment-btn"]').first()).toBeVisible()
    await deleteFlag(flag.id)
  })

  test('can update segment description and rollout', async ({ page }) => {
    const flag = await createFlag({ description: `seg-upd-${Date.now()}` })
    await createSegment(flag.id, `init-${Date.now()}`)
    await page.goto(`/#/flags/${flag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.locator('input[data-testid="segment-desc-input"]').first().fill(`upd-${Date.now()}`)
    await page.locator('input[data-testid="segment-rollout-input"]').first().fill('75')
    await page.dispatchEvent('[data-testid="save-segment-btn"]', 'click')
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    await deleteFlag(flag.id)
  })

  test('can delete a segment via UI', async ({ page }) => {
    const flag = await createFlag({ description: `seg-del-${Date.now()}` })
    await createSegment(flag.id, `del-${Date.now()}`)
    await page.goto(`/#/flags/${flag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.dispatchEvent('[data-testid="delete-segment-btn"]', 'click')
    await expect(page.locator('.el-message-box')).toBeVisible({ timeout: 5000 })
    await page.locator('.el-message-box .el-button--primary').click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    await deleteFlag(flag.id)
  })

  test('existing constraints are displayed', async ({ page }) => {
    const flag = await createFlag({ description: `cstr-display-${Date.now()}` })
    const seg = await createSegment(flag.id, `cstr-${Date.now()}`)
    await createConstraint(flag.id, seg.id)
    await page.goto(`/#/flags/${flag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.waitForSelector('input[data-testid="constraint-prop-input"]', { timeout: 5000 })
    await expect(page.locator('[data-testid="save-constraint-btn"]').first()).toBeVisible()
    await expect(page.locator('[data-testid="delete-constraint-btn"]').first()).toBeVisible()
    await deleteFlag(flag.id)
  })

  test('can create constraint via UI', async ({ page }) => {
    const flag = await createFlag({ description: `cstr-cr-${Date.now()}` })
    await createSegment(flag.id, `cstr-ui-${Date.now()}`)
    await page.goto(`/#/flags/${flag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.locator('input[data-testid="new-constraint-prop-input"]').first().fill('region')
    await page.locator('input[data-testid="new-constraint-value-input"]').first().fill('"EU"')
    await page.dispatchEvent('[data-testid="add-constraint-btn"]', 'click')
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    await deleteFlag(flag.id)
  })

  test('can update constraint property', async ({ page }) => {
    const flag = await createFlag({ description: `cstr-upd-${Date.now()}` })
    const seg = await createSegment(flag.id, `cstr-upd-${Date.now()}`)
    await createConstraint(flag.id, seg.id)
    await page.goto(`/#/flags/${flag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.waitForSelector('input[data-testid="constraint-prop-input"]', { timeout: 5000 })
    await page.locator('input[data-testid="constraint-prop-input"]').first().fill('region')
    await page.dispatchEvent('[data-testid="save-constraint-btn"]', 'click')
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    await deleteFlag(flag.id)
  })

  test('can delete constraint via UI', async ({ page }) => {
    const flag = await createFlag({ description: `cstr-del-${Date.now()}` })
    const seg = await createSegment(flag.id, `cstr-del-${Date.now()}`)
    await createConstraint(flag.id, seg.id)
    await page.goto(`/#/flags/${flag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.waitForSelector('[data-testid="delete-constraint-btn"]', { timeout: 5000 })
    await page.dispatchEvent('[data-testid="delete-constraint-btn"]', 'click')
    await expect(page.locator('.el-message-box')).toBeVisible({ timeout: 5000 })
    await page.locator('.el-message-box .el-button--primary').click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    await deleteFlag(flag.id)
  })

  test('delete flag via API succeeds', async ({ page }) => {
    const flag = await createFlag({ description: `del-${Date.now()}` })
    const resp = await fetch(`${API}/flags/${flag.id}`, { method: 'DELETE' })
    expect(resp.ok).toBeTruthy()
    const getResp = await fetch(`${API}/flags/${flag.id}`)
    expect(getResp.status).toBe(404)
  })

  test('history tab shows snapshots after changes', async ({ page }) => {
    await fetch(`${API}/flags/${testFlag.id}`, {
      method: 'PUT', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        description: `hist-${Date.now()}`,
        key: testFlag.key || '',
        dataRecordsEnabled: false,
        entityType: '',
        notes: ''
      })
    })
    await sleep(500)
    await page.goto(`/#/flags/${testFlag.id}`)
    await page.waitForSelector('input[data-testid="flag-key-input"]', { timeout: 10000 })
    await page.locator('.el-tabs__item').filter({ hasText: 'History' }).click()
    await sleep(2000)
    const cards = page.locator('.snapshot-container')
    const count = await cards.count()
    expect(count).toBeGreaterThan(0)
  })
})
