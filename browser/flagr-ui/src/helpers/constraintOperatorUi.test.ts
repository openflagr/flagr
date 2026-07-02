import { describe, expect, it } from 'vitest'
import {
  getOperatorHelpText,
  operatorApiBadge,
  operatorOptionDisplayText,
  operatorSelectClosedBadge,
  operatorSelectLabel,
} from './constraintOperatorUi'
import { findOperatorUi } from './constraintOperators'

describe('constraintOperatorUi', () => {
  it('getOperatorHelpText explains list vs substring for CONTAINS', () => {
    const line = getOperatorHelpText('CONTAINS')
    expect(line).toMatch(/roles includes/i)
    expect(line).toMatch(/Not text substring/i)
  })

  it('getOperatorHelpText uses same =~ framing for text includes and EREG', () => {
    const simple = getOperatorHelpText('UI_STRING_CONTAINS')
    const pattern = getOperatorHelpText('EREG')
    expect(simple).toMatch(/String property =~ value/)
    expect(pattern).toMatch(/String property =~ value/)
    expect(simple).toMatch(/plain text/i)
    expect(pattern).toMatch(/regex pattern/i)
  })

  it('getOperatorHelpText documents quoted strings for EQ', () => {
    expect(getOperatorHelpText('EQ')).toMatch(/quote/i)
    expect(getOperatorHelpText('EQ')).toMatch(/"US"/)
    expect(getOperatorHelpText('')).toBeNull()
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