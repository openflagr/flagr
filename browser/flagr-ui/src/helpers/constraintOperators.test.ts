import { describe, expect, it } from 'vitest'
import {
  OPERATOR_UI_OPTIONS,
  findOperatorUi,
  operatorOptionGroups,
} from './constraintOperators'

describe('constraintOperators', () => {
  it('exposes 12 API operators plus 2 UI sugar options', () => {
    expect(OPERATOR_UI_OPTIONS).toHaveLength(14)
    const values = OPERATOR_UI_OPTIONS.map((o) => o.value)
    expect(values).toContain('CONTAINS')
    expect(values).toContain('EREG')
    expect(values).toContain('UI_STRING_CONTAINS')
  })

  it('findOperatorUi resolves CONTAINS list semantics in description', () => {
    const op = findOperatorUi('CONTAINS')
    expect(op?.label).toBe('List includes')
    expect(op?.description).toMatch(/list/i)
    expect(op?.description).toMatch(/Not substring/i)
  })

  it('findOperatorUi resolves text includes sugar', () => {
    const op = findOperatorUi('UI_STRING_CONTAINS')
    expect(op?.label).toBe('Text includes')
    expect(op?.group).toBe('Text (simple)')
    expect(op?.valuePlaceholder).toBe('@gmail.com')
  })

  it('findOperatorUi resolves EREG substring guidance', () => {
    const op = findOperatorUi('EREG')
    expect(op?.description).toMatch(/substring|contains text/i)
    expect(op?.valuePlaceholder).toBe('"@gmail.com"')
  })

  it('operatorOptionGroups orders Compare, Lists, Text simple, Text pattern', () => {
    const groups = operatorOptionGroups()
    expect(groups.map((g) => g.label)).toEqual([
      'Compare',
      'Lists',
      'Text (simple)',
      'Text pattern',
    ])
    expect(groups[1].options.some((o) => o.value === 'IN')).toBe(true)
    expect(groups[2].options.some((o) => o.value === 'UI_STRING_CONTAINS')).toBe(true)
    expect(groups[3].options.some((o) => o.value === 'EREG')).toBe(true)
  })
})