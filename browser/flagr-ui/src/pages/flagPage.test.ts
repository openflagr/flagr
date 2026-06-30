import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { FlagPageVm } from './flagPage'
import { mountFlagPage } from './flagPage'

vi.mock('@/api/crud', () => ({
  loadFlagPageContext: vi.fn(() => new Promise(() => {})),
}))

function minimalVm(overrides: Partial<FlagPageVm> = {}): FlagPageVm {
  return {
    flagId: '42',
    flag: { description: '', tags: [], variants: [], segments: [] },
    loaded: true,
    historyLoaded: true,
    historyKey: 0,
    flagPageLoadGen: 0,
    flagSnapshots: [{ id: 1 }],
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
    const vm = minimalVm({ flagPageLoadGen: 3 })
    mountFlagPage(vm)

    expect(vm.flagPageLoadGen).toBe(4)
    expect(vm.loaded).toBe(false)
    expect(vm.historyLoaded).toBe(false)
    expect(vm.historyKey).toBe(1)
    expect(vm.flagSnapshots).toEqual([])
    expect(vm.dialogDuplicateFlagVisible).toBe(false)
    expect(vm.dialogEditDistributionOpen).toBe(false)
    expect(vm.dialogCreateSegmentOpen).toBe(false)
    expect(vm.selectedSegment).toBeNull()
    expect(vm.duplicateInFlight).toBe(false)
  })
})