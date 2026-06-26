import type { Effect } from 'effect'
import type { ApiError } from '@/api/errors'
import { runApi, type RunApiOptions, type RunApiVm } from './runApi'

export interface ConfirmVm extends RunApiVm {
  $confirm: (
    message: string,
    title: string,
    options: {
      confirmButtonText: string
      cancelButtonText: string
      type: 'warning'
    },
  ) => Promise<void>
}

export function confirmAndRunApi<A, E extends ApiError>(
  vm: ConfirmVm,
  confirmMessage: string,
  program: Effect.Effect<A, E>,
  options: RunApiOptions<A> = {},
): void {
  vm.$confirm(confirmMessage, 'Warning', {
    confirmButtonText: 'OK',
    cancelButtonText: 'Cancel',
    type: 'warning',
  })
    .then(() => runApi(vm, program, options))
    .catch(() => {})
}