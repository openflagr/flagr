import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiDecodeError, ApiHttpError, ApiUnauthorized } from './errors'
import { requestJson } from './http'

describe('requestJson', () => {
  const originalFetch = globalThis.fetch

  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
  })

  afterEach(() => {
    globalThis.fetch = originalFetch
    vi.unstubAllGlobals()
  })

  it('returns undefined on 204 without parsing JSON', async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response(null, { status: 204, statusText: 'No Content' }),
    )
    const result = await requestJson<void>({ method: 'DELETE', path: '/flags/1' })
    expect(result.ok).toBe(true)
    if (result.ok) expect(result.value).toBeUndefined()
  })

  it('maps 401 to ApiUnauthorized with redirect from WWW-Authenticate', async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response(JSON.stringify({ message: 'nope' }), {
        status: 401,
        headers: { 'WWW-Authenticate': 'redirect="https://auth.example/login"' },
      }),
    )
    const result = await requestJson({ method: 'GET', path: '/flags' })
    expect(result.ok).toBe(false)
    if (!result.ok) {
      expect(result.error).toBeInstanceOf(ApiUnauthorized)
      expect((result.error as ApiUnauthorized).redirectURL).toBe('https://auth.example/login')
    }
  })

  it('maps invalid JSON body to ApiDecodeError', async () => {
    vi.mocked(fetch).mockResolvedValue(new Response('not-json{', { status: 200 }))
    const result = await requestJson({ method: 'GET', path: '/flags' })
    expect(result.ok).toBe(false)
    if (!result.ok) expect(result.error).toBeInstanceOf(ApiDecodeError)
  })

  it('maps 500 with message to ApiHttpError', async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response(JSON.stringify({ message: 'server broke' }), { status: 500 }),
    )
    const result = await requestJson({ method: 'GET', path: '/flags' })
    expect(result.ok).toBe(false)
    if (!result.ok) {
      const err = result.error as ApiHttpError
      expect(err).toBeInstanceOf(ApiHttpError)
      expect(err.status).toBe(500)
      expect(err.message).toBe('server broke')
    }
  })
})