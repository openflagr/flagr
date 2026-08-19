import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import {
  getFlag,
  listAllTags,
  listDeletedFlags,
  listEntityTypes,
  listFlagSnapshots,
  listFlagsIfStale,
} from './crud'
import { clearDumpCache } from './evalCache'
import { evalOnlyMode } from '@/helpers/serverMode'

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

describe('eval-only mode reads', () => {
  const originalFetch = globalThis.fetch

  const exportedFlags = [
    { ID: 1, Key: 'f1', Description: 'a', Enabled: true, Tags: [{ ID: 10, Value: 't1' }] },
    { ID: 2, Key: 'f2', Description: 'b', Enabled: false, Tags: [{ ID: 20, Value: 't1' }] },
  ]

  beforeEach(() => {
    evalOnlyMode.value = true
    clearDumpCache()
    vi.stubGlobal('fetch', vi.fn())
    vi.mocked(fetch).mockImplementation((input) => {
      const url = String(input)
      if (url.includes('/export/eval_cache/json')) {
        return Promise.resolve(
          new Response(JSON.stringify({ Flags: exportedFlags }), {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          }),
        )
      }
      return Promise.reject(new Error(`unexpected fetch: ${url}`))
    })
  })

  afterEach(() => {
    evalOnlyMode.value = false
    globalThis.fetch = originalFetch
    vi.unstubAllGlobals()
  })

  it('listFlagsIfStale serves reversed flags from the export dump', async () => {
    const result = await listFlagsIfStale(undefined)
    expect(result.ok).toBe(true)
    if (result.ok && result.value) {
      expect(result.value.maxSnapshotID).toBe(0)
      expect(result.value.flags.map((f) => f.id)).toEqual([2, 1])
    }
  })

  it('listFlagsIfStale refetches even when a cached max id is passed', async () => {
    const first = await listFlagsIfStale(undefined)
    expect(first.ok).toBe(true)
    const second = await listFlagsIfStale(0)
    expect(second.ok).toBe(true)
    if (second.ok) expect(second.value).not.toBeNull()
    expect(vi.mocked(fetch)).toHaveBeenCalledTimes(2)
  })

  it('getFlag and listAllTags read from the export dump', async () => {
    const flagRes = await getFlag(2)
    expect(flagRes.ok).toBe(true)
    if (flagRes.ok) expect(flagRes.value.key).toBe('f2')

    const tagsRes = await listAllTags()
    expect(tagsRes.ok).toBe(true)
    if (tagsRes.ok) expect(tagsRes.value).toEqual([{ id: 10, value: 't1' }])
  })

  it('snapshots, entity types, and deleted flags resolve empty without network', async () => {
    const [snapshots, entityTypes, deleted] = await Promise.all([
      listFlagSnapshots(1),
      listEntityTypes(),
      listDeletedFlags(),
    ])
    expect(snapshots.ok && snapshots.value).toEqual([])
    expect(entityTypes.ok && entityTypes.value).toEqual([])
    expect(deleted.ok && deleted.value).toEqual([])
    expect(vi.mocked(fetch)).not.toHaveBeenCalled()
  })
})