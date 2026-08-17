import { test, expect, type Page } from '@playwright/test'

/**
 * Read-only (eval-only) mode UI. The shared e2e backend runs in writable
 * mode, so the two endpoints the UI reads in this mode are intercepted:
 * /health flips the mode and /export/eval_cache/json serves a GitOps-shaped
 * dump (PascalCase entity.Flag). Backend 403 enforcement is covered by
 * TestEvalOnlyDenyMiddleware in pkg/config.
 */

const exportDump = {
  Flags: [
    {
      ID: 1,
      UpdatedAt: '2026-01-02T03:04:05Z',
      Key: 'readonly_demo_flag',
      Description: 'readonly e2e demo flag',
      CreatedBy: '',
      UpdatedBy: '',
      Enabled: true,
      Notes: '',
      DataRecordsEnabled: false,
      EntityType: '',
      Tags: [{ ID: 1, Value: 'readonly-e2e' }],
      Variants: [
        { ID: 1, FlagID: 1, Key: 'control', Attachment: null },
        { ID: 2, FlagID: 1, Key: 'treatment', Attachment: { color: 'blue' } },
      ],
      Segments: [
        {
          ID: 1,
          FlagID: 1,
          Description: 'everyone',
          Rank: 0,
          RolloutPercent: 100,
          Constraints: [
            { ID: 1, SegmentID: 1, Property: 'env', Operator: 'EQ', Value: '"prod"' },
          ],
          Distributions: [
            { ID: 1, SegmentID: 1, VariantID: 1, VariantKey: 'control', Percent: 100 },
          ],
        },
      ],
    },
  ],
}

async function interceptEvalOnly(page: Page): Promise<void> {
  await page.route('**/api/v1/health', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ status: 'OK', evalOnlyMode: true }),
    }),
  )
  await page.route('**/api/v1/export/eval_cache/json', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(exportDump),
    }),
  )
}

test.describe('read-only (eval-only) mode', () => {
  test.beforeEach(async ({ page }) => {
    await interceptEvalOnly(page)
  })

  test('flags list shows the banner and hides write affordances', async ({ page }) => {
    await page.goto('/')

    await expect(page.locator('[data-testid="readonly-banner"]')).toBeVisible()
    await expect(page.getByText('readonly e2e demo flag')).toBeVisible()

    await expect(page.locator('[data-testid="create-flag-btn"]')).toHaveCount(0)
    await expect(page.getByText('Deleted Flags')).toHaveCount(0)
  })

  test('flag detail is browsable but not editable; Debug Console stays', async ({ page }) => {
    await page.goto('/#/flags/1')

    await expect(page.locator('input[data-testid="flag-key-input"]')).toHaveValue(
      'readonly_demo_flag',
    )
    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeDisabled()
    await expect(page.locator('input[data-testid="flag-desc-input"]')).toBeDisabled()

    await expect(page.locator('[data-testid="save-flag-btn"]')).toHaveCount(0)
    await expect(page.locator('[data-testid="delete-flag-btn"]')).toHaveCount(0)
    await expect(page.locator('[data-testid="duplicate-flag-btn"]')).toHaveCount(0)
    await expect(page.locator('[data-testid="create-segment-btn"]')).toHaveCount(0)

    // Data from the dump renders (segment, variants, tag) — read-only inputs.
    await expect(page.locator('input[data-testid="segment-desc-input"]')).toHaveValue('everyone')
    await expect(page.locator('input[data-testid="segment-desc-input"]')).toBeDisabled()
    await expect(page.locator('input[data-testid="variant-key-input"]').first()).toHaveValue(
      'control',
    )
    await expect(page.getByText('readonly-e2e')).toBeVisible()

    // History tab is gone; Debug Console remains.
    await expect(page.locator('#tab-history')).toHaveCount(0)
    await expect(page.locator('.dc-container')).toBeVisible()
  })

  test('history deep link lands on the Config tab', async ({ page }) => {
    await page.goto('/#/flags/1?tab=history')

    await expect(page.locator('input[data-testid="flag-key-input"]')).toBeVisible()
    await expect(page.locator('#tab-history')).toHaveCount(0)
    await expect(page.locator('#tab-config')).toHaveClass(/is-active/)
  })
})
