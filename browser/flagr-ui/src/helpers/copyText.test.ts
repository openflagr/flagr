import { afterEach, describe, expect, it, vi } from 'vitest'
import { COPY_FEEDBACK_MS, SNAPSHOT_HIGHLIGHT_MS, copyText } from './copyText'

describe('copyText', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
  })

  it('exposes feedback timing constants', () => {
    expect(COPY_FEEDBACK_MS).toBe(2000)
    expect(SNAPSHOT_HIGHLIGHT_MS).toBe(1500)
  })

  it('uses navigator.clipboard.writeText when available', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    vi.stubGlobal('navigator', { clipboard: { writeText } })

    await expect(copyText('https://example.com/#/flags/1')).resolves.toBe(true)
    expect(writeText).toHaveBeenCalledWith('https://example.com/#/flags/1')
  })

  it('falls back when clipboard API rejects', async () => {
    const writeText = vi.fn().mockRejectedValue(new Error('denied'))
    vi.stubGlobal('navigator', { clipboard: { writeText } })

    const execCommand = vi.fn().mockReturnValue(true)
    const appendChild = vi.fn()
    const removeChild = vi.fn()
    const ta = {
      value: '',
      setAttribute: vi.fn(),
      style: {} as CSSStyleDeclaration,
      focus: vi.fn(),
      select: vi.fn(),
    }
    vi.stubGlobal('document', {
      createElement: vi.fn(() => ta),
      body: { appendChild, removeChild },
      execCommand,
    })

    await expect(copyText('hello')).resolves.toBe(true)
    expect(ta.value).toBe('hello')
    expect(execCommand).toHaveBeenCalledWith('copy')
  })

  it('returns false when both paths fail', async () => {
    vi.stubGlobal('navigator', { clipboard: { writeText: vi.fn().mockRejectedValue(new Error('x')) } })
    vi.stubGlobal('document', {
      createElement: vi.fn(() => {
        throw new Error('no dom')
      }),
      body: { appendChild: vi.fn(), removeChild: vi.fn() },
      execCommand: vi.fn(),
    })

    await expect(copyText('x')).resolves.toBe(false)
  })
})
