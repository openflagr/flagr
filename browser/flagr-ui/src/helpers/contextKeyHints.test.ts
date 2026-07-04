import { describe, expect, it } from 'vitest'
import { contextKeyHint } from './contextKeyHints'

describe('contextKeyHint', () => {
  describe('@ts (Unix epoch seconds)', () => {
    it('formats epoch 0 as Jan 1, 1970', () => {
      const result = contextKeyHint('@ts', '0')
      expect(result).toContain('Jan 1, 1970')
      expect(result).toContain('UTC')
    })

    it('formats a future timestamp', () => {
      // Just verify it formats a valid timestamp with UTC
      const result = contextKeyHint('@ts', '1751666400')
      expect(result).toMatch(/UTC$/)
      expect(result).toContain('=')
    })

    it('returns null for non-numeric value', () => {
      expect(contextKeyHint('@ts', 'abc')).toBeNull()
    })

    it('returns null for empty value', () => {
      expect(contextKeyHint('@ts', '')).toBeNull()
    })
  })

  describe('@ts_hour (0-23 UTC)', () => {
    it('0 = 12:00 AM (midnight)', () => {
      expect(contextKeyHint('@ts_hour', '0')).toBe('= 00:00 UTC (12 AM)')
    })

    it('9 = 9:00 AM', () => {
      expect(contextKeyHint('@ts_hour', '9')).toBe('= 09:00 UTC (9 AM)')
    })

    it('12 = 12:00 PM (noon)', () => {
      expect(contextKeyHint('@ts_hour', '12')).toBe('= 12:00 UTC (12 PM)')
    })

    it('14 = 2:00 PM', () => {
      expect(contextKeyHint('@ts_hour', '14')).toBe('= 14:00 UTC (2 PM)')
    })

    it('23 = 11:00 PM', () => {
      expect(contextKeyHint('@ts_hour', '23')).toBe('= 23:00 UTC (11 PM)')
    })

    it('returns null for out-of-range value', () => {
      expect(contextKeyHint('@ts_hour', '24')).toBeNull()
      expect(contextKeyHint('@ts_hour', '-1')).toBeNull()
    })
  })

  describe('@ts_weekday (0=Sunday, 6=Saturday)', () => {
    it('0 = Sunday', () => {
      expect(contextKeyHint('@ts_weekday', '0')).toBe('= Sunday')
    })

    it('1 = Monday', () => {
      expect(contextKeyHint('@ts_weekday', '1')).toBe('= Monday')
    })

    it('5 = Friday', () => {
      expect(contextKeyHint('@ts_weekday', '5')).toBe('= Friday')
    })

    it('6 = Saturday', () => {
      expect(contextKeyHint('@ts_weekday', '6')).toBe('= Saturday')
    })

    it('returns null for out-of-range value', () => {
      expect(contextKeyHint('@ts_weekday', '7')).toBeNull()
      expect(contextKeyHint('@ts_weekday', '-1')).toBeNull()
    })
  })

  describe('@ts_month (1-12)', () => {
    it('1 = January', () => {
      expect(contextKeyHint('@ts_month', '1')).toBe('= January')
    })

    it('7 = July', () => {
      expect(contextKeyHint('@ts_month', '7')).toBe('= July')
    })

    it('12 = December', () => {
      expect(contextKeyHint('@ts_month', '12')).toBe('= December')
    })

    it('returns null for out-of-range value', () => {
      expect(contextKeyHint('@ts_month', '0')).toBeNull()
      expect(contextKeyHint('@ts_month', '13')).toBeNull()
    })
  })

  describe('non-ts properties', () => {
    it('returns null for non-ts properties', () => {
      expect(contextKeyHint('country', 'US')).toBeNull()
      expect(contextKeyHint('tier', 'premium')).toBeNull()
    })
  })
})
