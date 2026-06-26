import { Match } from 'effect'
import type { ApiError } from '@/api/errors'
import { ensureApiError } from '@/api/errors'

export interface MessageApi {
  error: (msg: string) => void
}

/** User-visible message for a tagged API failure (exhaustive on `ApiError`). */
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