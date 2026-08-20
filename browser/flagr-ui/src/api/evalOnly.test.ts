import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { clearDumpCache, fetchFlags, getFlag, getFlags, listAllTags, mapExportedFlag } from './evalOnly'
import { ApiHttpError } from './errors'

/** PascalCase export fixture mirroring pkg/handler/testdata/sample_eval_cache.json. */
const exportedFlag = {
  ID: 1,
  CreatedAt: '2018-10-05T23:10:25.234392-07:00',
  UpdatedAt: '2019-04-15T23:02:08.034382-07:00',
  DeletedAt: null,
  Key: 'kmmcd1nsd6',
  Description: 'demo_example',
  CreatedBy: 'alice',
  UpdatedBy: 'bob',
  Enabled: true,
  Notes: 'some notes',
  DataRecordsEnabled: true,
  EntityType: 'user',
  SnapshotID: 7,
  Tags: [
    { ID: 10, Value: 'team-a', CreatedAt: '2019-01-01T00:00:00Z', DeletedAt: null },
  ],
  Variants: [
    { ID: 2, FlagID: 1, Key: 'blue123', Attachment: { color: 'blue' }, DeletedAt: null },
    { ID: 3, FlagID: 1, Key: 'red', Attachment: null, DeletedAt: null },
  ],
  Segments: [
    {
      ID: 6,
      FlagID: 1,
      Description: 'Users in CA',
      Rank: 0,
      RolloutPercent: 100,
      DeletedAt: null,
      Constraints: [
        { ID: 3, SegmentID: 6, Property: 'env', Operator: 'EQ', Value: '"local"', DeletedAt: null },
      ],
      Distributions: [
        { ID: 7, SegmentID: 6, VariantID: 2, VariantKey: 'blue123', Percent: 100, DeletedAt: null },
      ],
    },
  ],
}

describe('mapExportedFlag', () => {
  it('maps the exported PascalCase shape to the swagger camelCase Flag', () => {
    const flag = mapExportedFlag(exportedFlag)
    expect(flag).toEqual({
      id: 1,
      key: 'kmmcd1nsd6',
      description: 'demo_example',
      enabled: true,
      notes: 'some notes',
      createdBy: 'alice',
      updatedBy: 'bob',
      updatedAt: '2019-04-15T23:02:08.034382-07:00',
      dataRecordsEnabled: true,
      entityType: 'user',
      tags: [{ id: 10, value: 'team-a' }],
      variants: [
        { id: 2, key: 'blue123', attachment: { color: 'blue' } },
        { id: 3, key: 'red', attachment: undefined },
      ],
      segments: [
        {
          id: 6,
          description: 'Users in CA',
          rank: 0,
          rolloutPercent: 100,
          constraints: [{ id: 3, property: 'env', operator: 'EQ', value: '"local"' }],
          distributions: [{ id: 7, percent: 100, variantID: 2, variantKey: 'blue123' }],
        },
      ],
    })
  })

  it('materializes empty arrays for null nested collections', () => {
    const flag = mapExportedFlag({
      ID: 5,
      Key: 'k',
      Description: 'd',
      Enabled: false,
      Tags: null,
      Variants: null,
      Segments: null,
    })
    expect(flag.tags).toEqual([])
    expect(flag.variants).toEqual([])
    expect(flag.segments).toEqual([])
  })
})

describe('eval-only adapter', () => {
  const originalFetch = globalThis.fetch

  beforeEach(() => {
    clearDumpCache()
    vi.stubGlobal('fetch', vi.fn())
  })

  afterEach(() => {
    globalThis.fetch = originalFetch
    vi.unstubAllGlobals()
  })

  function stubExport(flags: unknown[], status = 200) {
    vi.mocked(fetch).mockImplementation((input) => {
      const url = String(input)
      if (url.includes('/export/eval_cache/json')) {
        return Promise.resolve(
          new Response(JSON.stringify({ Flags: flags }), {
            status,
            headers: { 'Content-Type': 'application/json' },
          }),
        )
      }
      return Promise.reject(new Error(`unexpected fetch: ${url}`))
    })
  }

  it('fetchFlags sorts by id regardless of export order', async () => {
    stubExport([
      { ...exportedFlag, ID: 3, Key: 'c' },
      { ...exportedFlag, ID: 1, Key: 'a' },
      { ...exportedFlag, ID: 2, Key: 'b' },
    ])
    const res = await fetchFlags()
    expect(res.ok).toBe(true)
    if (res.ok) expect(res.value.map((f) => f.id)).toEqual([1, 2, 3])
  })

  it('getFlags reads through the cache without another fetch', async () => {
    stubExport([exportedFlag])
    await fetchFlags()
    const res = await getFlags()
    expect(res.ok).toBe(true)
    expect(vi.mocked(fetch)).toHaveBeenCalledTimes(1)
  })

  it('concurrent reads on an empty cache share one request', async () => {
    stubExport([exportedFlag])
    const [flagRes, tagsRes] = await Promise.all([getFlag(1), listAllTags()])
    expect(flagRes.ok).toBe(true)
    expect(tagsRes.ok).toBe(true)
    expect(vi.mocked(fetch)).toHaveBeenCalledTimes(1)
  })

  it('getFlag returns a 404-shaped error for an unknown id', async () => {
    stubExport([exportedFlag])
    const res = await getFlag(999)
    expect(res.ok).toBe(false)
    if (!res.ok) {
      expect(res.error).toBeInstanceOf(ApiHttpError)
      expect((res.error as ApiHttpError).status).toBe(404)
    }
  })

  it('listAllTags dedupes tags by value across flags', async () => {
    stubExport([
      { ...exportedFlag, ID: 1, Tags: [{ ID: 10, Value: 'team-a' }] },
      {
        ...exportedFlag,
        ID: 2,
        Tags: [
          { ID: 20, Value: 'team-a' },
          { ID: 21, Value: 'team-b' },
        ],
      },
    ])
    const res = await listAllTags()
    expect(res.ok).toBe(true)
    if (res.ok) {
      expect(res.value).toEqual([
        { id: 10, value: 'team-a' },
        { id: 21, value: 'team-b' },
      ])
    }
  })

  it('propagates http errors from the export endpoint', async () => {
    stubExport([], 500)
    const res = await fetchFlags()
    expect(res.ok).toBe(false)
  })
})
