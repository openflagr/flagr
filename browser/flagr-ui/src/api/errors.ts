export class ApiHttpError {
  readonly _tag = 'ApiHttpError' as const
  constructor(
    readonly status: number,
    readonly message: string,
  ) {}
}

export class ApiUnauthorized {
  readonly _tag = 'ApiUnauthorized' as const
  constructor(readonly redirectURL?: string) {}
}

export class ApiNetworkError {
  readonly _tag = 'ApiNetworkError' as const
  constructor(readonly cause: unknown) {}
}

export class ApiDecodeError {
  readonly _tag = 'ApiDecodeError' as const
  constructor(readonly cause: unknown) {}
}

export type ApiError =
  | ApiHttpError
  | ApiUnauthorized
  | ApiNetworkError
  | ApiDecodeError

export function isApiError(u: unknown): u is ApiError {
  return (
    u instanceof ApiHttpError ||
    u instanceof ApiUnauthorized ||
    u instanceof ApiNetworkError ||
    u instanceof ApiDecodeError
  )
}

export function ensureApiError(cause: unknown): ApiError {
  return isApiError(cause) ? cause : new ApiNetworkError(cause)
}