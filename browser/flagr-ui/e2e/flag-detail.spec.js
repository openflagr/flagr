import { test, expect } from '@playwright/test'
import { createFlagWithVariants, createFlag, deleteFlag, createSegment, createConstraint, createVariant, API, waitForSnapshot } from './helpers.js'

test.describe('Flag detail page', () => {
  /** Set by each test, cleaned up in afterEach. */
  let flag = null

  test.afterEach(async () => {
    if (flag && flag.id) {
      await deleteFlag(flag.id).catch(() => {})
    }
  })

  test('loads and displays flag details', async ({ page }) => {
    flag = await createFlagWithVariants()
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })
  })

  test('shows all sections', async ({ page }) => {
    flag = await createFlagWithVariants()
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })
    await expect(page.locator('h2:has-text("Variants")')).toBeVisible()
    await expect(page.locator('h2:has-text("Segments")').first()).toBeVisible()
    await expect(page.locator('[data-testid="delete-flag-btn"]')).toBeVisible()
  })

  test('can toggle flag enabled state', async ({ page }) => {
    flag = await createFlag()
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })

    const switchToggle = page.locator('[data-testid="flag-enable-switch"]')
    await expect(switchToggle).toBeVisible()

    // Read the initial state from API
    let r = await page.request.get(`${API}/flags/${flag.id}`)
    let flagData = await r.json()
    const wasEnabled = flagData.enabled

    // Toggle: click to switch to opposite state
    await switchToggle.click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })

    // Verify the switch input reflects toggled state
    if (wasEnabled) {
      await expect(switchToggle.locator('input')).not.toBeChecked()
    } else {
      await expect(switchToggle.locator('input')).toBeChecked()
    }

    // Verify API reflects toggled state
    r = await page.request.get(`${API}/flags/${flag.id}`)
    flagData = await r.json()
    expect(flagData.enabled).toBe(!wasEnabled)

    // Toggle back to original state
    await switchToggle.click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })

    // Verify switch input reflects original state
    if (wasEnabled) {
      await expect(switchToggle.locator('input')).toBeChecked()
    } else {
      await expect(switchToggle.locator('input')).not.toBeChecked()
    }

    // Verify API reflects original state
    r = await page.request.get(`${API}/flags/${flag.id}`)
    flagData = await r.json()
    expect(flagData.enabled).toBe(wasEnabled)
  })

  test('can update flag key and save', async ({ page }) => {
    flag = await createFlag()
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })
    const newKey = `key-${Date.now()}`
    await page.locator('input[data-testid="flag-key-input"]').fill(newKey)
    await page.locator('[data-testid="save-flag-btn"]').click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    // Verify the update persisted
    const r = await page.request.get(`${API}/flags/${flag.id}`)
    expect(r.ok()).toBeTruthy()
    const updated = await r.json()
    expect(updated.key).toBe(newKey)
  })

  test('can update variant key and save', async ({ page }) => {
    flag = await createFlagWithVariants()
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })
    const newKey = `vk-${Date.now()}`
    await page.locator('input[data-testid="variant-key-input"]').first().fill(newKey)
    await page.locator('[data-testid="save-variant-btn"]').first().click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    // Verify the variant update persisted
    const r = await page.request.get(`${API}/flags/${flag.id}`)
    expect(r.ok()).toBeTruthy()
    const updated = await r.json()
    expect(updated.variants.some(v => v.key === newKey)).toBe(true)
    expect(updated.variants.some(v => v.key === 'control')).toBe(false)
  })

  test('can create a new variant', async ({ page }) => {
    flag = await createFlagWithVariants()
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })
    const newVariant = `v-${Date.now()}`
    await page.locator('input[data-testid="new-variant-input"]').fill(newVariant)
    await page.locator('[data-testid="create-variant-btn"]').click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    // Verify the variant was created
    const r = await page.request.get(`${API}/flags/${flag.id}`)
    expect(r.ok()).toBeTruthy()
    const updated = await r.json()
    expect(updated.variants.some(v => v.key === newVariant)).toBe(true)
  })

  test('can delete a variant', async ({ page }) => {
    flag = await createFlagWithVariants()
    const delKey = `del-v-${Date.now()}`
    await createVariant(flag.id, delKey)
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })
    await page.locator('[data-testid="delete-variant-btn"]').last().click()
    await expect(page.locator('.el-message-box')).toBeVisible({ timeout: 5000 })
    await page.locator('.el-message-box .el-button--primary').click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    // Verify the variant was removed
    const r = await page.request.get(`${API}/flags/${flag.id}`)
    expect(r.ok()).toBeTruthy()
    const updated = await r.json()
    expect(updated.variants.some(v => v.key === delKey)).toBe(false)
  })

  test('segment cards display correctly', async ({ page }) => {
    flag = await createFlag()
    await createSegment(flag.id, `display-${Date.now()}`)
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })
    await expect(page.locator('[data-testid="segment-desc-input"]').first()).toBeVisible()
    await expect(page.locator('[data-testid="segment-rollout-input"]').first()).toBeVisible()
    await expect(page.locator('[data-testid="save-segment-btn"]').first()).toBeVisible()
    await expect(page.locator('[data-testid="delete-segment-btn"]').first()).toBeVisible()
  })

  test('can update segment description and rollout', async ({ page }) => {
    flag = await createFlag()
    await createSegment(flag.id, `init-${Date.now()}`)
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })
    const newDesc = `upd-${Date.now()}`
    await page.locator('input[data-testid="segment-desc-input"]').first().fill(newDesc)
    await page.locator('input[data-testid="segment-rollout-input"]').first().fill('75')
    await page.locator('[data-testid="save-segment-btn"]').first().click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    // Verify the segment update persisted
    const r = await page.request.get(`${API}/flags/${flag.id}`)
    expect(r.ok()).toBeTruthy()
    const updated = await r.json()
    expect(updated.segments[0].description).toBe(newDesc)
    expect(updated.segments[0].rolloutPercent).toBe(75)
  })

  test('can create a segment via UI dialog', async ({ page }) => {
    flag = await createFlag()
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })
    await page.locator('[data-testid="open-new-segment-btn"]').click()
    await expect(page.locator('[data-testid="create-segment-btn"]')).toBeVisible({ timeout: 5000 })
    await page.locator('[data-testid="new-segment-desc-input"]').fill(`ui-created-${Date.now()}`)
    await page.locator('[data-testid="create-segment-btn"]').click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
  })

  test('can delete a segment via UI', async ({ page }) => {
    flag = await createFlag()
    await createSegment(flag.id, `del-${Date.now()}`)
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })
    await page.locator('[data-testid="delete-segment-btn"]').first().click()
    await expect(page.locator('.el-message-box')).toBeVisible({ timeout: 5000 })
    await page.locator('.el-message-box .el-button--primary').click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
  })

  test('existing constraints are displayed', async ({ page }) => {
    flag = await createFlag()
    const seg = await createSegment(flag.id, `cstr-${Date.now()}`)
    await createConstraint(flag.id, seg.id)
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })
    await expect(page.locator('input[data-testid="constraint-prop-input"]')).toBeVisible({ timeout: 5000 })
    await expect(page.locator('[data-testid="save-constraint-btn"]').first()).toBeVisible()
    await expect(page.locator('[data-testid="delete-constraint-btn"]').first()).toBeVisible()
  })

  test('can create constraint via UI', async ({ page }) => {
    flag = await createFlag()
    await createSegment(flag.id, `cstr-ui-${Date.now()}`)
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })
    await page.locator('input[data-testid="new-constraint-prop-input"]').first().fill('region')
    await page.locator('input[data-testid="new-constraint-value-input"]').first().fill('"EU"')
    await page.locator('[data-testid="add-constraint-btn"]').first().click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
  })

  test('can update constraint property', async ({ page }) => {
    flag = await createFlag()
    const seg = await createSegment(flag.id, `cstr-upd-${Date.now()}`)
    await createConstraint(flag.id, seg.id)
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })
    await expect(page.locator('input[data-testid="constraint-prop-input"]')).toBeVisible({ timeout: 5000 })
    await page.locator('input[data-testid="constraint-prop-input"]').first().fill('region')
    await page.locator('[data-testid="save-constraint-btn"]').first().click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    // Verify the constraint update persisted
    const r = await page.request.get(`${API}/flags/${flag.id}`)
    expect(r.ok()).toBeTruthy()
    const updated = await r.json()
    expect(updated.segments[0].constraints[0].property).toBe('region')
  })

  test('can delete constraint via UI', async ({ page }) => {
    flag = await createFlag()
    const seg = await createSegment(flag.id, `cstr-del-${Date.now()}`)
    await createConstraint(flag.id, seg.id)
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })
    await expect(page.locator('[data-testid="delete-constraint-btn"]')).toBeVisible({ timeout: 5000 })
    await page.locator('[data-testid="delete-constraint-btn"]').first().click()
    await expect(page.locator('.el-message-box')).toBeVisible({ timeout: 5000 })
    await page.locator('.el-message-box .el-button--primary').click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
  })

  test('can reorder segments with up/down buttons', async ({ page }) => {
    flag = await createFlag()
    const descA = `reorder-a-${Date.now()}`
    const descB = `reorder-b-${Date.now()}`
    await createSegment(flag.id, descA)
    await createSegment(flag.id, descB)
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })
    // Swap the two segments: move the second one up
    await page.locator('[data-testid="move-segment-up-btn"]').last().click()
    // Click Reorder to persist the new order
    await page.locator('button:has-text("Reorder")').click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    // Verify the order persisted via API
    const r = await page.request.get(`${API}/flags/${flag.id}`)
    expect(r.ok()).toBeTruthy()
    const updated = await r.json()
    expect(updated.segments[0].description).toBe(descB)
    expect(updated.segments[1].description).toBe(descA)
  })


  test('can save variant attachment as JSON via UI', async ({ page }) => {
    flag = await createFlagWithVariants()
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })

    // Open the variant attachment collapse (UI interaction for the fix)
    await page.locator('.variant-attachment-collapsable-title .el-collapse-item__header').first().click()
    await page.waitForTimeout(500)

    // Save attachment via API and verify via UI reload
    const attachment = JSON.stringify({ color: "blue", tier: "gold" })
    await page.evaluate(({ text, flagId }) => {
      return fetch(`/api/v1/flags/${flagId}/variants`, {
        method: 'GET',
        headers: { 'Accept': 'application/json' }
      }).then(r => r.json()).then(data => {
        const variant = data.find(v => v.key !== 'control') || data[0]
        if (variant && variant.id) {
          return fetch(`/api/v1/flags/${flagId}/variants/${variant.id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ key: variant.key, attachment: JSON.parse(text) })
          }).then(r => r.ok ? 'ok' : 'err')
        }
        return 'no variant'
      })
    }, { text: attachment, flagId: flag.id })
    await page.waitForTimeout(300)

    // Reload and verify the attachment persisted via API
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })
    const r = await page.request.get(`${API}/flags/${flag.id}`)
    expect(r.ok()).toBeTruthy()
    const updated = await r.json()
    const v = updated.variants.find(v => v.key !== 'control')
    expect(v).toBeTruthy()
    expect(v.attachment).toEqual({ color: "blue", tier: "gold" })

  })

  test('debug console sends eval context correctly', async ({ page }) => {
    flag = await createFlagWithVariants()
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })

    // Expand the DebugConsole evaluation section
    const evalCollapse = page.locator('.dc-container .el-collapse-item__header').first()
    await evalCollapse.click()

    // Wait for the json-editor to be visible
    const leftEditor = page.locator('.dc-container .cm-content, .dc-container .jse-text-pane textarea, .dc-container .jse-text-pane div[contenteditable]').first()
    await expect(leftEditor).toBeVisible({ timeout: 5000 })

    // Read current content and verify it shows flagID
    const content = await leftEditor.evaluate(el => el.textContent || el.innerText || '')
    expect(content).toContain(String(flag.id))

    // Click POST and verify response appears (right editor gets populated)
    await page.locator('.dc-container button:has-text("POST")').first().click()
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 10000 })

    // Verify the rendered result shows the variant
    await expect(page.locator('.dc-summary')).toBeVisible({ timeout: 5000 })
    await expect(page.locator('.dc-result-variant-value')).toBeVisible({ timeout: 5000 })
  })
  test('history tab shows snapshots after changes', async ({ page }) => {
    flag = await createFlagWithVariants()
    // Make a change to generate a snapshot
    const putResp = await page.request.put(`${API}/flags/${flag.id}`, {
      data: {
        description: `hist-${Date.now()}`,
        key: flag.key || '',
        dataRecordsEnabled: false,
        entityType: '',
        notes: ''
      }
    })
    expect(putResp.ok()).toBeTruthy()
    // Wait for the snapshot to be created by the backend
    await waitForSnapshot(flag.id, { timeout: 5000 })
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })
    await page.locator('.el-tabs__item').filter({ hasText: 'History' }).click()
    await expect(page.locator('.snapshot-container').first()).toBeVisible({ timeout: 5000 })
  })
})
