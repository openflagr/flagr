import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { FlagPageVm } from './flagPage'
import {
  applyDeepLink,
  mountFlagPage,
  scrollToSnapshot,
} from './flagPage'
import { SNAPSHOT_HIGHLIGHT_MS } from '@/helpers/copyText'
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
