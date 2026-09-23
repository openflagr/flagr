import { test, expect } from '@playwright/test'
import { API, createFlag, createSegment, deleteFlag } from './helpers'

test.describe('Context enrichers', () => {
  /** Set by each test, cleaned up in afterEach. */
  let flag = null

  test.afterEach(async () => {
    if (flag && flag.id) {
      await deleteFlag(flag.id).catch(() => {})
    }
  })

  test('adds a Jev enricher and saves an edited question', async ({ page }) => {
    flag = await createFlag()
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })

    // Regression: this used to POST a question with blank instructions, which the
    // server rejected with a 400, leaving the button looking broken.
    await page.locator('[data-testid="add-jev-enricher-btn"]').click()
    await expect(page.locator('.el-message--success:has-text("enricher created")')).toBeVisible({ timeout: 5000 })

    const row = page.locator('[data-testid="enricher-row-jev"]')
    await expect(row).toBeVisible()
    await expect(row.locator('[data-testid="jev-question-name"]')).toHaveValue('example_question')
    await expect(row.locator('[data-testid="jev-question-instructions"]')).not.toHaveValue('')

    // The editor teaches the mapping: property preview plus the value shape.
    await expect(row.locator('[data-testid="jev-property-preview"]')).toHaveText('@jev_example_question')
    await expect(row.locator('.jev-value-hint')).toContainText('P(true)')

    // Persisted, and its enriched property is in the effective catalog.
    let r = await page.request.get(`${API}/flags/${flag.id}`)
    let data = await r.json()
    let jev = data.enrichers.find((e) => e.namespace === 'jev')
    expect(jev).toBeTruthy()
    expect(jev.properties).toContain('@jev_example_question')

    // Rename, save, and verify the property follows.
    await row.locator('[data-testid="jev-question-name"]').fill('plan_tier')
    await expect(row.locator('[data-testid="jev-property-preview"]')).toHaveText('@jev_plan_tier')
    await expect(row.locator('[data-testid="save-enricher-jev"]')).toBeEnabled()
    await row.locator('[data-testid="save-enricher-jev"]').click()
    await expect(page.locator('.el-message--success:has-text("enricher saved")')).toBeVisible({ timeout: 5000 })

    r = await page.request.get(`${API}/flags/${flag.id}`)
    data = await r.json()
    jev = data.enrichers.find((e) => e.namespace === 'jev')
    expect(jev.properties).toContain('@jev_plan_tier')
    expect(jev.properties).not.toContain('@jev_example_question')
  })

  test('blocks save and explains an invalid question', async ({ page }) => {
    flag = await createFlag()
    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })

    await page.locator('[data-testid="add-jev-enricher-btn"]').click()
    await expect(page.locator('.el-message--success:has-text("enricher created")')).toBeVisible({ timeout: 5000 })

    const row = page.locator('[data-testid="enricher-row-jev"]')
    await expect(row).toBeVisible()

    await row.locator('[data-testid="jev-question-instructions"]').fill('')
    await expect(row.locator('[data-testid="save-enricher-jev"]')).toBeDisabled()
    await expect(row.locator('.enricher-problem')).toContainText('needs instructions')

    // The last question cannot be removed: a Jev enricher must ask something.
    await expect(row.locator('[data-testid="jev-remove-question-0"]')).toBeDisabled()
  })

  test('offers the enriched property in the constraint picker', async ({ page }) => {
    flag = await createFlag()
    await createSegment(flag.id, 'segment')

    const res = await page.request.post(`${API}/flags/${flag.id}/enrichers`, {
      data: {
        namespace: 'jev',
        config: {
          questions: {
            plan_tier: { type: 'noul', instructions: 'Is the plan enterprise?' },
          },
        },
      },
    })
    expect(res.ok()).toBeTruthy()

    await page.goto(`/#/flags/${flag.id}`)
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible({ timeout: 10000 })

    await page.locator('[data-testid="pick-enriched-property-btn"]').first().click()
    const option = page.locator('[data-testid="enriched-property-@jev_plan_tier"]')
    await expect(option).toBeVisible()
    await option.click()
    await expect(page.locator('input[data-testid="new-constraint-prop-input"]').first()).toHaveValue('@jev_plan_tier')
  })
})
