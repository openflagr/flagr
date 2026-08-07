import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { evalOnlyMode, initServerMode } from './serverMode'

describe('initServerMode', () => {
  const originalFetch = globalThis.fetch

  beforeEach(() => {
    evalOnlyMode.value = false
    vi.stubGlobal('fetch', vi.fn())
  })

  afterEach(() => {
    globalThis.fetch = originalFetch
    vi.unstubAllGlobals()
  })

  function jsonResponse(body: unknown, status = 200) {
    return new Response(JSON.stringify(body), {
      status,
      headers: { 'Content-Type': 'application/json' },
    })
  }

  it('sets evalOnlyMode when health reports it', async () => {
    vi.mocked(fetch).mockResolvedValue(jsonResponse({ status: 'OK', evalOnlyMode: true }))

    await initServerMode()
    expect(evalOnlyMode.value).toBe(true)
  })

  it('stays editable when health reports evalOnlyMode false', async () => {
    vi.mocked(fetch).mockResolvedValue(jsonResponse({ status: 'OK', evalOnlyMode: false }))

    await initServerMode()
    expect(evalOnlyMode.value).toBe(false)
  })

  it('stays editable for older servers without the field', async () => {
    vi.mocked(fetch).mockResolvedValue(jsonResponse({ status: 'OK' }))

    await initServerMode()
    expect(evalOnlyMode.value).toBe(false)
  })

  it('fails open when the health check errors', async () => {
    evalOnlyMode.value = true
    vi.mocked(fetch).mockRejectedValue(new Error('network down'))

    await initServerMode()
    expect(evalOnlyMode.value).toBe(false)
  })
})
