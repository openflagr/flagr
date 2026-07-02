import { describe, expect, it } from 'vitest'
import {
  UI_STRING_CONTAINS,
  UI_STRING_NOT_CONTAINS,
  escapeRegexLiteral,
  isStoredLiteralSubstringPattern,
  isUiSugarOperator,
  materializeConstraintForApi,
  plainTextFromStoredConstraintValue,
  resolveUiOperator,
} from './constraintOperatorSugar'

describe('constraintOperatorSugar', () => {
  it('isUiSugarOperator identifies sugar ids only', () => {
    expect(isUiSugarOperator(UI_STRING_CONTAINS)).toBe(true)
    expect(isUiSugarOperator(UI_STRING_NOT_CONTAINS)).toBe(true)
    expect(isUiSugarOperator('EQ')).toBe(false)
    expect(isUiSugarOperator('EREG')).toBe(false)
  })

  it('escapeRegexLiteral escapes regexp metacharacters', () => {
    expect(escapeRegexLiteral('@gmail.com')).toBe('@gmail\\.com')
    expect(escapeRegexLiteral('a+b')).toBe('a\\+b')
  })

  it('detects literal substring patterns', () => {
    expect(isStoredLiteralSubstringPattern('"premium"')).toBe(true)
    expect(isStoredLiteralSubstringPattern('".*foo.*"')).toBe(false)
  })

  it('resolveUiOperator maps literal EREG to string contains sugar', () => {
    expect(resolveUiOperator('EREG', '"premium"')).toBe(UI_STRING_CONTAINS)
    expect(resolveUiOperator('EREG', '".+@x"')).toBe('EREG')
  })

  it('materializeConstraintForApi maps UI string contains to EREG', () => {
    const out = materializeConstraintForApi({
      property: 'email',
      operator: UI_STRING_CONTAINS,
      value: '@gmail.com',
    })
    expect(out.operator).toBe('EREG')
    expect(out.value).toBe('"@gmail\\.com"')
  })

  it('materializeConstraintForApi maps UI string not contains to NEREG', () => {
    const out = materializeConstraintForApi({
      property: 'path',
      operator: UI_STRING_NOT_CONTAINS,
      value: '/admin',
    })
    expect(out.operator).toBe('NEREG')
    expect(out.value).toBe('"/admin"')
  })

  it('plainTextFromStoredConstraintValue round-trips literals', () => {
    expect(plainTextFromStoredConstraintValue('"premium"')).toBe('premium')
  })
})