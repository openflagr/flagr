import { describe, expect, it } from 'vitest'
import {
  operatorApiBadge,
  operatorHelpText,
  operatorOptionDisplayText,
  operatorSelectClosedBadge,
  operatorSelectLabel,
} from './constraintOperatorUi'
import { findOperatorUi } from './constraintOperators'

describe('constraintOperatorUi', () => {
  it('operatorHelpText returns catalog hint for option row', () => {
    const op = findOperatorUi('CONTAINS')!
    expect(operatorHelpText(op)).toMatch(/roles includes/i)
  })

  it('operatorHelpText explains list vs substring for CONTAINS', () => {
    const line = operatorHelpText(findOperatorUi('CONTAINS'))
    expect(line).toMatch(/roles includes/i)
    expect(line).toMatch(/Not text substring/i)
  })

  it('operatorHelpText uses same =~ framing for text includes and EREG', () => {
    const simple = operatorHelpText(findOperatorUi('UI_STRING_CONTAINS'))
    const pattern = operatorHelpText(findOperatorUi('EREG'))
    expect(simple).toMatch(/String property =~ value/)
    expect(pattern).toMatch(/String property =~ value/)
    expect(simple).toMatch(/plain text/i)
    expect(pattern).toMatch(/regex pattern/i)
  })

  it('operatorHelpText documents quoted strings for EQ', () => {
    expect(operatorHelpText(findOperatorUi('EQ'))).toMatch(/quote/i)
    expect(operatorHelpText(findOperatorUi('EQ'))).toMatch(/"US"/)
    expect(operatorHelpText(undefined)).toBeNull()
  })

  it('operatorHelpText states compare operators are numbers-only', () => {
    for (const op of ['LT', 'LTE', 'GT', 'GTE'] as const) {
      const line = operatorHelpText(findOperatorUi(op))
      expect(line).toMatch(/Numbers only/i)
      expect(line).toMatch(/No string/i)
      expect(line).toMatch(/Equals/i)
    }
  })

  it('operatorApiBadge maps API tokens', () => {
    expect(operatorApiBadge('EQ')).toBe('==')
    expect(operatorApiBadge('IN')).toBe('IN')
  })

  it('operatorSelectLabel uses human text for el-option label', () => {
    const eq = findOperatorUi('EQ')!
    expect(operatorSelectLabel(eq)).toBe('Equals')
  })

  it('operatorSelectClosedBadge shows API token when value set', () => {
    expect(operatorSelectClosedBadge('EQ')).toBe('==')
    expect(operatorSelectClosedBadge('')).toBe('')
  })

  it('operatorOptionDisplayText strips trailing syntax in parens', () => {
    const eq = findOperatorUi('EQ')!
    expect(operatorOptionDisplayText(eq)).toBe('Equals')
    expect(operatorOptionDisplayText({ ...eq, label: 'equals (==)' })).toBe('equals')
  })
})