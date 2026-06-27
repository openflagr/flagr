import constants from '@/helpers/constants'
import {
  ApiDecodeError,
  ApiHttpError,
  ApiNetworkError,
  ApiUnauthorized,
  ensureApiError,
} from './errors'
import type { ApiError } from './errors'
import type { ApiResult } from './result'
import { err, ok } from './result'

const { API_URL } = constants

type Method = 'GET' | 'POST' | 'PUT' | 'DELETE'

export interface RequestOptions {
  readonly method: Method
  readonly path: string
  readonly body?: unknown
}

async function parseErrorMessage(res: Response): Promise<string> {
  try {
    const data = (await res.json()) as { message?: string }
    if (typeof data?.message === 'string') return data.message
  } catch {
    /* ignore */
  }
  return 'request error'
}

/** Map non-OK responses; 401 uses first quoted URL in WWW-Authenticate (Flagr JWT middleware). */
function mapResponseError(res: Response, message: string): ApiError {
  if (res.status === 401) {
    const auth = res.headers.get('www-authenticate') ?? ''
    const match = auth.split('"')[1]
    return new ApiUnauthorized(match)
  }
  return new ApiHttpError(res.status, message)
}

async function decodeJsonBody<T>(res: Response): Promise<ApiResult<T>> {
  if (res.status === 204 || res.status === 205) {
    return ok(undefined as T)
  }
  let text: string
  try {
    text = await res.text()
  } catch (cause) {
    return err(new ApiNetworkError(cause))
  }
  if (!text) {
    return ok(undefined as T)
  }
  try {
    return ok(JSON.parse(text) as T)
  } catch (cause) {
    return err(new ApiDecodeError(cause))
  }
}

export async function requestJson<T>(opts: RequestOptions): Promise<ApiResult<T>> {
  const url = `${API_URL}${opts.path}`
  const init: RequestInit = {
    method: opts.method,
    headers: opts.body !== undefined ? { 'Content-Type': 'application/json' } : undefined,
    body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
  }

  let res: Response
  try {
    res = await fetch(url, init)
  } catch (cause) {
    return err(ensureApiError(cause))
  }

  if (!res.ok) {
    let message: string
    try {
      message = await parseErrorMessage(res)
    } catch (cause) {
      return err(new ApiNetworkError(cause))
    }
    return err(mapResponseError(res, message))
  }

  return decodeJsonBody<T>(res)
}

/** Mutations with empty or omitted JSON bodies. */
export async function requestVoid(opts: RequestOptions): Promise<ApiResult<void>> {
  return requestJson<void>(opts)
}