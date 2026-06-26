import { Cause, Effect, Exit, Match } from 'effect'
import type { ApiError } from '@/api/errors'
import { ensureApiError } from '@/api/errors'

export interface MessageApi {
  error: (msg: string) => void
}

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
 */
export interface RunApiOptions<A> {
  onSuccess?: (value: A) => void
  successMessage?: string
}

export function apiErrorUserMessage(error: ApiError): string {
  return Match.valueTags(error, {
    ApiHttpError: (e) => e.message,
    ApiUnauthorized: () => 'request error',
    ApiNetworkError: () => 'request error',
    ApiDecodeError: () => 'request error',
  })
}

export function presentApiError(error: ApiError, message: MessageApi): void {
  if (error._tag === 'ApiUnauthorized' && error.redirectURL) {
    message.error(apiErrorUserMessage(error))
    window.location.href = error.redirectURL
    return
  }
  message.error(apiErrorUserMessage(error))
}

export function presentUnknownFailure(cause: unknown, message: MessageApi): void {
  presentApiError(ensureApiError(cause), message)
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
