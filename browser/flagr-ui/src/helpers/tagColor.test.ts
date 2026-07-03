import { describe, expect, it } from 'vitest'
import { tagColor } from './tagColor'

describe('tagColor', () => {
  it('returns same colour for same input', () => {
    expect(tagColor('team:payments')).toBe(tagColor('team:payments'))
  })

  it('returns different colours for different inputs', () => {
    expect(tagColor('team:payments')).not.toBe(tagColor('team:auth'))
  })

  it('returns valid HSL string', () => {
    const color = tagColor('test')
    expect(color).toMatch(/^hsl\(\d+, 55%, 92%\)$/)
  })

  it('handles empty string', () => {
    expect(tagColor('')).toBe('hsl(0, 55%, 92%)')
  })
})
