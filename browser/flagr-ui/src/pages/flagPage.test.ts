import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { FlagPageVm } from './flagPage'
import {
  applyDeepLink,
  loadFlagSnapshots,
  loadOlderFlagSnapshots,
  mountFlagPage,
  scrollToSnapshot,
} from './flagPage'
import * as crudApi from '@/api/crud'
import { ok } from '@/api/result'
import type { ApiResult } from '@/api/result'
import type { FlagSnapshot } from '@/api/types'
import { SNAPSHOT_HIGHLIGHT_MS } from '@/helpers/copyText'
import { evalOnlyMode, snapshotsHistoryPageSize } from '@/helpers/serverMode'
import { FLAG_TAB_CONFIG, FLAG_TAB_HISTORY, snapshotElementId } from '@/helpers/shareLinks'

vi.mock('@/api/crud', () => ({
  loadFlagPageContext: vi.fn(() => new Promise(() => {})),
  listFlagSnapshots: vi.fn(() => new Promise(() => {})),
}))

function minimalVm(overrides: Partial<FlagPageVm> = {}): FlagPageVm {
  return {
    flagId: '42',
    flag: { description: '', tags: [], variants: [], segments: [] },
    loaded: true,
    activeTab: FLAG_TAB_CONFIG,
    historyLoaded: true,
    historyKey: 0,
    flagPageLoadGen: 0,
    flagSnapshots: [{ id: 1 }],
    pendingSnapshotScrollId: null,
    dialogDuplicateFlagVisible: true,
    dialogEditDistributionOpen: true,
    dialogCreateSegmentOpen: true,
    selectedSegment: { id: 1 } as FlagPageVm['selectedSegment'],
    $message: Object.assign(vi.fn(), { error: vi.fn(), success: vi.fn(), warning: vi.fn() }),
    $confirm: vi.fn(),
    $router: { replace: vi.fn() },
    evalContext: { entityID: '', entityType: '', entityContext: {}, enableDebug: false },
    batchEvalContext: { entities: [], enableDebug: false, flagIDs: [] },
    evalResult: {},
    batchEvalResult: { evaluationResults: [] },
    ...overrides,
  } as FlagPageVm
}

describe('mountFlagPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('resets route-local state and bumps flagPageLoadGen', () => {
    const vm = minimalVm({
      flagPageLoadGen: 3,
      activeTab: FLAG_TAB_HISTORY,
      pendingSnapshotScrollId: 9,
    })
    mountFlagPage(vm)

    expect(vm.flagPageLoadGen).toBe(4)
    expect(vm.loaded).toBe(false)
    expect(vm.activeTab).toBe(FLAG_TAB_CONFIG)
    expect(vm.historyLoaded).toBe(false)
    expect(vm.historyKey).toBe(1)
    expect(vm.flagSnapshots).toEqual([])
    expect(vm.historyHasMore).toBe(false)
    expect(vm.historyLoadingOlder).toBe(false)
    expect(vm.pendingSnapshotScrollId).toBeNull()
    expect(vm.dialogDuplicateFlagVisible).toBe(false)
    expect(vm.dialogEditDistributionOpen).toBe(false)
    expect(vm.dialogCreateSegmentOpen).toBe(false)
    expect(vm.selectedSegment).toBeNull()
    expect(vm.duplicateInFlight).toBe(false)
  })
})

describe('applyDeepLink', () => {
  it('opens history and queues snapshot scroll', () => {
    const vm = minimalVm({ historyKey: 2, historyLoaded: false })
    applyDeepLink(vm, { tab: 'history', snapshot: '87' })

    expect(vm.activeTab).toBe(FLAG_TAB_HISTORY)
    expect(vm.historyLoaded).toBe(true)
    expect(vm.historyKey).toBe(3)
    expect(vm.pendingSnapshotScrollId).toBe(87)
  })

  it('resets to config when query has no history intent', () => {
    const vm = minimalVm({
      activeTab: FLAG_TAB_HISTORY,
      pendingSnapshotScrollId: 3,
    })
    applyDeepLink(vm, {})

    expect(vm.activeTab).toBe(FLAG_TAB_CONFIG)
    expect(vm.pendingSnapshotScrollId).toBeNull()
  })

  it('routes a history deep link to config in read-only mode', () => {
    evalOnlyMode.value = true
    try {
      const vm = minimalVm({ historyKey: 2, historyLoaded: false })
      applyDeepLink(vm, { tab: 'history', snapshot: '87' })

      expect(vm.activeTab).toBe(FLAG_TAB_CONFIG)
      expect(vm.historyLoaded).toBe(false)
      expect(vm.pendingSnapshotScrollId).toBeNull()
    } finally {
      evalOnlyMode.value = false
    }
  })
})

describe('scrollToSnapshot', () => {
  afterEach(() => {
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  it('scrolls, highlights, and clears highlight after timeout', () => {
    vi.useFakeTimers()
    const el = {
      id: snapshotElementId(5),
      classList: {
        tokens: new Set<string>(),
        add(c: string) {
          this.tokens.add(c)
        },
        remove(c: string) {
          this.tokens.delete(c)
        },
        contains(c: string) {
          return this.tokens.has(c)
        },
      },
      scrollIntoView: vi.fn(),
    }
    vi.stubGlobal('document', {
      getElementById: (id: string) => (id === el.id ? el : null),
    })
    vi.stubGlobal('window', {
      matchMedia: () => ({ matches: true }),
      setTimeout: globalThis.setTimeout.bind(globalThis),
    })

    expect(scrollToSnapshot(5)).toBe(true)
    expect(el.scrollIntoView).toHaveBeenCalled()
    expect(el.classList.contains('snapshot-container--highlight')).toBe(true)

    vi.advanceTimersByTime(SNAPSHOT_HIGHLIGHT_MS)
    expect(el.classList.contains('snapshot-container--highlight')).toBe(false)
  })

  it('returns false when the snapshot node is missing', () => {
    vi.stubGlobal('document', { getElementById: () => null })
    expect(scrollToSnapshot(999)).toBe(false)
  })
})

describe('snapshot paging', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    snapshotsHistoryPageSize.value = 0
  })

  afterEach(() => {
    snapshotsHistoryPageSize.value = 0
  })

  function fakeSnapshots(n: number, startId: number): FlagSnapshot[] {
    return Array.from({ length: n }, (_, i) => ({
      id: startId - i,
      flag: {},
      updatedAt: '2026-09-22T00:00:00Z',
    })) as FlagSnapshot[]
  }

  it('fetches the whole history and shows no button when the server limit is 0', async () => {
    snapshotsHistoryPageSize.value = 0
    const vm = minimalVm({ flagSnapshots: [], historyHasMore: true, historyLoadingOlder: false })
    vi.mocked(crudApi.listFlagSnapshots).mockResolvedValue(ok(fakeSnapshots(130, 130)))

    loadFlagSnapshots(vm)
    await vi.waitFor(() => expect(vm.flagSnapshots).toHaveLength(130))
    // undefined page => full-history request
    expect(crudApi.listFlagSnapshots).toHaveBeenCalledWith('42', undefined)
    expect(vm.historyHasMore).toBe(false)
  })

  it('loadFlagSnapshots fetches the first page and flags a longer history', async () => {
    snapshotsHistoryPageSize.value = 50
    const vm = minimalVm({ flagSnapshots: [], historyHasMore: false, historyLoadingOlder: false })
    vi.mocked(crudApi.listFlagSnapshots).mockResolvedValue(ok(fakeSnapshots(50, 100)))

    loadFlagSnapshots(vm)
    await vi.waitFor(() => expect(vm.flagSnapshots).toHaveLength(50))
    expect(crudApi.listFlagSnapshots).toHaveBeenCalledWith('42', { limit: 50, offset: 0 })
    expect(vm.historyHasMore).toBe(true)
  })

  it('loadFlagSnapshots keeps hasMore off for a short history', async () => {
    snapshotsHistoryPageSize.value = 50
    const vm = minimalVm({ flagSnapshots: [], historyHasMore: true, historyLoadingOlder: false })
    vi.mocked(crudApi.listFlagSnapshots).mockResolvedValue(ok(fakeSnapshots(3, 3)))

    loadFlagSnapshots(vm)
    await vi.waitFor(() => expect(vm.flagSnapshots).toHaveLength(3))
    expect(vm.historyHasMore).toBe(false)
  })

  it('loadOlderFlagSnapshots appends the next page at the current offset', async () => {
    snapshotsHistoryPageSize.value = 50
    const vm = minimalVm({
      flagSnapshots: fakeSnapshots(50, 100),
      historyHasMore: true,
      historyLoadingOlder: false,
    })
    vi.mocked(crudApi.listFlagSnapshots).mockResolvedValue(ok(fakeSnapshots(10, 50)))

    loadOlderFlagSnapshots(vm)
    await vi.waitFor(() => expect(vm.flagSnapshots).toHaveLength(60))
    expect(crudApi.listFlagSnapshots).toHaveBeenCalledWith('42', { limit: 50, offset: 50 })
    // A short page means the history is exhausted.
    expect(vm.historyHasMore).toBe(false)
    expect(vm.historyLoadingOlder).toBe(false)
  })

  it('loadOlderFlagSnapshots is a no-op while loading, when nothing is left, or unpaginated', () => {
    snapshotsHistoryPageSize.value = 50
    const loading = minimalVm({ historyHasMore: true, historyLoadingOlder: true })
    loadOlderFlagSnapshots(loading)
    const exhausted = minimalVm({ historyHasMore: false, historyLoadingOlder: false })
    loadOlderFlagSnapshots(exhausted)
    snapshotsHistoryPageSize.value = 0
    const unpaged = minimalVm({ historyHasMore: true, historyLoadingOlder: false })
    loadOlderFlagSnapshots(unpaged)
    expect(crudApi.listFlagSnapshots).not.toHaveBeenCalled()
  })

  it('loadOlderFlagSnapshots drops a stale page after the list was reloaded', async () => {
    snapshotsHistoryPageSize.value = 50
    const vm = minimalVm({
      flagSnapshots: fakeSnapshots(50, 100),
      historyHasMore: true,
      historyLoadingOlder: false,
    })
    let resolvePage: (v: ApiResult<FlagSnapshot[]>) => void = () => {}
    vi.mocked(crudApi.listFlagSnapshots).mockReturnValue(
      new Promise<ApiResult<FlagSnapshot[]>>((resolve) => {
        resolvePage = resolve
      }),
    )

    loadOlderFlagSnapshots(vm)
    // History reloads (e.g. the tab is reopened) while the page is in flight.
    vm.historyKey++
    vm.flagSnapshots = fakeSnapshots(50, 200)
    resolvePage(ok(fakeSnapshots(50, 50)))

    await vi.waitFor(() => expect(vm.historyLoadingOlder).toBe(false))
    expect(vm.flagSnapshots).toHaveLength(50)
  })
})
