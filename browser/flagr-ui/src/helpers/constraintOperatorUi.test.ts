import { describe, expect, it } from 'vitest'
import { getOperatorHintLine, operatorApiBadge, operatorOptionDisplayText, operatorSelectClosedBadge, operatorSelectLabel } from './constraintOperatorUi'
import { findOperatorUi } from './constraintOperators'

describe('constraintOperatorUi', () => {
  it('getOperatorHintLine explains list vs substring for CONTAINS', () => {
    const line = getOperatorHintLine('CONTAINS')
    expect(line).toMatch(/roles includes/i)
    expect(line).toMatch(/Not text substring/i)
  })

  it('getOperatorHintLine is null for EQ', () => {
    expect(getOperatorHintLine('EQ')).toBeNull()
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