import { Effect } from 'effect'
import constants from '@/constants'
import {
  ApiDecodeError,
  ApiHttpError,
  ApiNetworkError,
  ApiUnauthorized,
  ensureApiError,
  type ApiError,
} from './errors'

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

function mapResponseError(res: Response, message: string): ApiError {
  if (res.status === 401) {
    const auth = res.headers.get('www-authenticate') ?? ''
    const match = auth.split('"')[1]
    return new ApiUnauthorized({ redirectURL: match })
  }
  return new ApiHttpError({ status: res.status, message })
}

const decodeJsonBody = Effect.fn('flagr.decodeJsonBody')(function* <T>(res: Response) {
  if (res.status === 204 || res.status === 205) {
    return undefined as T
  }
  const text = yield* Effect.tryPromise({
    try: () => res.text(),
    catch: (cause) => new ApiNetworkError({ cause }),
  })
  if (!text) {
    return undefined as T
  }
  try {
    return JSON.parse(text) as T
  } catch (cause) {
    return yield* Effect.fail(new ApiDecodeError({ cause }))
  }
})

export const requestJson = Effect.fn('flagr.requestJson')(function* <T>(opts: RequestOptions) {
  const url = `${API_URL}${opts.path}`
  const init: RequestInit = {
    method: opts.method,
    headers: opts.body !== undefined ? { 'Content-Type': 'application/json' } : undefined,
    body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
  }

  const res = yield* Effect.tryPromise({
    try: () => fetch(url, init),
    catch: (cause) => ensureApiError(cause),
  })

  if (!res.ok) {
    const message = yield* Effect.tryPromise({
      try: () => parseErrorMessage(res),
      catch: (cause) => new ApiNetworkError({ cause }),
    })
    return yield* Effect.fail(mapResponseError(res, message))
  }

  return yield* decodeJsonBody<T>(res)
})

/** Mutations with empty or omitted JSON bodies. */
export const requestVoid = (opts: RequestOptions): Effect.Effect<void, ApiError> =>
  requestJson<void>(opts)