import type { ApiError } from '@/api/errors'
import { ensureApiError } from '@/api/errors'
import type { ApiResult } from '@/api/result'

export interface MessageApi {
  error: (msg: string) => void
}

export interface ElementMessageApi extends MessageApi {
  success: (msg: string) => void
  warning: (msg: string) => void
  (options: {
    type?: 'success' | 'warning' | 'info' | 'error'
    message?: string
    duration?: number
    showClose?: boolean
    dangerouslyUseHTMLString?: boolean
    customClass?: string
  }): void
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
  onFailure?: () => void
  successMessage?: string
}

export function apiErrorUserMessage(error: ApiError): string {
  switch (error._tag) {
    case 'ApiHttpError':
      return error.message
    case 'ApiUnauthorized':
    case 'ApiNetworkError':
    case 'ApiDecodeError':
      return 'request error'
    default: {
      const _exhaustive: never = error
      return _exhaustive
    }
  }
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

export function runApi<A>(
  vm: RunApiVm,
  promise: Promise<ApiResult<A>>,
  options: RunApiOptions<A> = {},
): void {
  void promise.then((result) => {
    if (!result.ok) {
      options.onFailure?.()
      presentApiError(result.error, vm.$message)
      return
    }
    if (options.successMessage) {
      vm.$message.success(options.successMessage)
    }
    options.onSuccess?.(result.value)
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

export function confirmAndRunApi<A>(
  vm: ConfirmVm,
  confirmMessage: string,
  promise: Promise<ApiResult<A>>,
  options: RunApiOptions<A> = {},
): void {
  vm.$confirm(confirmMessage, 'Warning', {
    confirmButtonText: 'OK',
    cancelButtonText: 'Cancel',
    type: 'warning',
  })
    .then(() => runApi(vm, promise, options))
    .catch(() => {})
}