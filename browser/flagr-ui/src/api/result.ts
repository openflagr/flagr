import type { ApiError } from './errors'

export type ApiResult<T> = { ok: true; value: T } | { ok: false; error: ApiError }

export function ok<T>(value: T): ApiResult<T> {
  return { ok: true, value }
}

export function err<T>(error: ApiError): ApiResult<T> {
  return { ok: false, error }
}