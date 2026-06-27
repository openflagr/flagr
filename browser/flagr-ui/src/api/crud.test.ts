import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { listFlagsIfStale } from './crud'

describe('listFlagsIfStale', () => {
  const originalFetch = globalThis.fetch

  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
  })

  afterEach(() => {
    globalThis.fetch = originalFetch
    vi.unstubAllGlobals()
  })

  function jsonResponse(body: unknown, status = 200) {
    return new Response(JSON.stringify(body), {
      status,
      headers: { 'Content-Type': 'application/json' },
    })
  }

  it('returns null when cached max id matches server', async () => {
    vi.mocked(fetch).mockImplementation((input) => {
      const url = String(input)
      if (url.includes('/flags/snapshots/max_id')) {
        return Promise.resolve(jsonResponse({ maxID: 42 }))
      }
      return Promise.reject(new Error(`unexpected fetch: ${url}`))
    })

    const result = await listFlagsIfStale(42)
    expect(result.ok).toBe(true)
    if (result.ok) expect(result.value).toBeNull()
  })

  it('returns reversed flags when cache is stale', async () => {
    vi.mocked(fetch).mockImplementation((input) => {
      const url = String(input)
      if (url.includes('/flags/snapshots/max_id')) {
        return Promise.resolve(jsonResponse({ maxID: 2 }))
      }
      if (url.endsWith('/flags') || url.includes('/api/v1/flags')) {
        return Promise.resolve(
          jsonResponse([
            { id: 1, description: 'a', variants: [] },
            { id: 2, description: 'b', variants: [] },
          ]),
        )
      }
      return Promise.reject(new Error(`unexpected fetch: ${url}`))
    })

    const result = await listFlagsIfStale(1)
    expect(result.ok).toBe(true)
    if (result.ok && result.value) {
      expect(result.value.maxSnapshotID).toBe(2)
      expect(result.value.flags.map((f) => f.id)).toEqual([2, 1])
    }
  })
})