import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { getFlagReadSource, setFlagReadSource } from '@/api/crud'
import { evalOnlyMode, initServerMode } from './serverMode'

describe('initServerMode', () => {
  const originalFetch = globalThis.fetch

  beforeEach(() => {
    evalOnlyMode.value = false
    setFlagReadSource('http')
    vi.stubGlobal('fetch', vi.fn())
  })

  afterEach(() => {
    evalOnlyMode.value = false
    setFlagReadSource('http')
    globalThis.fetch = originalFetch
    vi.unstubAllGlobals()
  })

  function jsonResponse(body: unknown, status = 200) {
    return new Response(JSON.stringify(body), {
      status,
      headers: { 'Content-Type': 'application/json' },
    })
  }

  it('sets evalOnlyMode and the eval-cache read source when health reports it', async () => {
    vi.mocked(fetch).mockResolvedValue(jsonResponse({ status: 'OK', evalOnlyMode: true }))

    await initServerMode()
    expect(evalOnlyMode.value).toBe(true)
    expect(getFlagReadSource()).toBe('evalCache')
  })

  it('stays editable when health reports evalOnlyMode false', async () => {
    vi.mocked(fetch).mockResolvedValue(jsonResponse({ status: 'OK', evalOnlyMode: false }))

    await initServerMode()
    expect(evalOnlyMode.value).toBe(false)
    expect(getFlagReadSource()).toBe('http')
  })

  it('stays editable for older servers without the field', async () => {
    vi.mocked(fetch).mockResolvedValue(jsonResponse({ status: 'OK' }))

    await initServerMode()
    expect(evalOnlyMode.value).toBe(false)
    expect(getFlagReadSource()).toBe('http')
  })

  it('fails open when the health check errors', async () => {
    evalOnlyMode.value = true
    setFlagReadSource('evalCache')
    vi.mocked(fetch).mockRejectedValue(new Error('network down'))

    await initServerMode()
    expect(evalOnlyMode.value).toBe(false)
    expect(getFlagReadSource()).toBe('http')
  })

  it('fails open when the health wait is aborted', async () => {
    const ac = new AbortController()
    ac.abort()
    vi.mocked(fetch).mockImplementation((_input, init) => {
      if (init?.signal?.aborted) {
        return Promise.reject(new DOMException('The operation was aborted.', 'AbortError'))
      }
      return Promise.resolve(jsonResponse({ status: 'OK', evalOnlyMode: true }))
    })

    await initServerMode(ac.signal)
    expect(evalOnlyMode.value).toBe(false)
    expect(getFlagReadSource()).toBe('http')
  })
})
