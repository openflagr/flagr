import { describe, expect, it, vi } from 'vitest'
import { ApiHttpError } from '@/api/errors'
import { err, ok } from '@/api/result'
import type { RunApiVm } from './runApi'
import { runApi } from './runApi'

function testVm(overrides: Partial<RunApiVm> = {}): RunApiVm {
  return {
    $message: Object.assign(vi.fn(), {
      error: vi.fn(),
      success: vi.fn(),
      warning: vi.fn(),
    }),
    ...overrides,
  } as RunApiVm
}

describe('runApi', () => {
  it('calls onFailure and presents error when result is not ok', async () => {
    const vm = testVm()
    const onFailure = vi.fn()

    runApi(vm, Promise.resolve(err(new ApiHttpError(400, 'cannot duplicate flag'))), { onFailure })

    await vi.waitFor(() => {
      expect(onFailure).toHaveBeenCalledOnce()
      expect(vm.$message.error).toHaveBeenCalledWith('cannot duplicate flag')
    })
  })

  it('calls onSuccess when result is ok', async () => {
    const onSuccess = vi.fn()
    const vm = testVm()

    runApi(vm, Promise.resolve(ok({ id: 1 })), { onSuccess })

    await vi.waitFor(() => {
      expect(onSuccess).toHaveBeenCalledWith({ id: 1 })
    })
  })
})