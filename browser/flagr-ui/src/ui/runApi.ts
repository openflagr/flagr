import { Cause, Effect, Exit } from 'effect'
import type { ApiError } from '@/api/errors'
import { presentUnknownFailure, type MessageApi } from './presentApiError'

export interface ElementMessageApi extends MessageApi {
  success: (msg: string) => void
  warning: (msg: string) => void
}

export interface RunApiVm {
  $message: ElementMessageApi
}

/**
 * - `successMessage`: static toast after success (preferred for fixed copy).
 * - `onSuccess`: state updates only; use a dynamic toast here only when the message depends on the result.
 * Do not set both for the same success path.
 */
export interface RunApiOptions<A> {
  onSuccess?: (value: A) => void
  successMessage?: string
}

export function runApi<A, E extends ApiError>(
  vm: RunApiVm,
  program: Effect.Effect<A, E>,
  options: RunApiOptions<A> = {},
): void {
  void Effect.runPromiseExit(program).then((exit) => {
    Exit.match(exit, {
      onSuccess: (value) => {
        if (options.successMessage) {
          vm.$message.success(options.successMessage)
        }
        options.onSuccess?.(value)
      },
      onFailure: (cause) => {
        presentUnknownFailure(Cause.squash(cause), vm.$message)
      },
    })
  })
}