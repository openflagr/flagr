import { Data } from 'effect'

export class ApiHttpError extends Data.TaggedError('ApiHttpError')<{
  readonly status: number
  readonly message: string
}> {}

export class ApiUnauthorized extends Data.TaggedError('ApiUnauthorized')<{
  readonly redirectURL?: string
}> {}

export class ApiNetworkError extends Data.TaggedError('ApiNetworkError')<{
  readonly cause: unknown
}> {}

export class ApiDecodeError extends Data.TaggedError('ApiDecodeError')<{
  readonly cause: unknown
}> {}

export type ApiError =
  | ApiHttpError
  | ApiUnauthorized
  | ApiNetworkError
  | ApiDecodeError

/** Narrow unknown failures from `Cause.squash` / `tryPromise` to the API error union. */
export function isApiError(u: unknown): u is ApiError {
  return (
    u instanceof ApiHttpError ||
    u instanceof ApiUnauthorized ||
    u instanceof ApiNetworkError ||
    u instanceof ApiDecodeError
  )
}

export function ensureApiError(cause: unknown): ApiError {
  return isApiError(cause) ? cause : new ApiNetworkError({ cause })
}