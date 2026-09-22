import { describe, expect, it } from 'vitest'
import {
  choiceCriteriaFromRows,
  choiceRowsFromCriteria,
  defaultJevQuestion,
  isJevQuestionReady,
  jevPropertyFor,
  jevPropertyName,
  nextChoiceName,
  noulCriteriaText,
  scoreLevelsFromCriteria,
  slugifyJevName,
  withNoulCriteria,
} from './jevQuestion'

describe('jevQuestion', () => {
  it('maps between a property and its question name', () => {
    expect(jevPropertyName('@jev.buying_intent')).toBe('buying_intent')
    expect(jevPropertyName('dl_state')).toBe('')
    expect(jevPropertyFor('buying intent')).toBe('@jev.buying_intent')
    expect(slugifyJevName('risk-score v2')).toBe('risk_score_v2')
  })

  it('builds type-appropriate default questions', () => {
    expect(defaultJevQuestion('noul')).toEqual({ type: 'noul', instructions: '' })
    expect(defaultJevQuestion('choice')).toEqual({ type: 'choice', instructions: '', criteria: {} })
    expect(defaultJevQuestion('score')).toEqual({
      type: 'score',
      instructions: '',
      criteria: ['', ''],
    })
  })

  it('round-trips choice criteria through editable rows', () => {
    const criteria = { free: 'self-serve', pro: 'card payment' }
    const rows = choiceRowsFromCriteria(criteria)
    expect(rows).toEqual([
      { name: 'free', description: 'self-serve' },
      { name: 'pro', description: 'card payment' },
    ])
    expect(choiceCriteriaFromRows(rows)).toEqual(criteria)
  })

  it('ignores non-object choice criteria', () => {
    expect(choiceRowsFromCriteria(undefined)).toEqual([])
    expect(choiceRowsFromCriteria(['a', 'b'])).toEqual([])
  })

  it('maps score criteria to level strings', () => {
    expect(scoreLevelsFromCriteria(['low', 'medium', 'high'])).toEqual(['low', 'medium', 'high'])
    expect(scoreLevelsFromCriteria(undefined)).toEqual([])
    expect(scoreLevelsFromCriteria({ low: 'x' })).toEqual([])
  })

  it('reads and writes noul criteria', () => {
    expect(noulCriteriaText({ true: 'mentions a prior attempt' }, 'true')).toBe(
      'mentions a prior attempt',
    )
    expect(noulCriteriaText(undefined, 'false')).toBe('')
    expect(withNoulCriteria(undefined, 'true', 'yes case')).toEqual({ true: 'yes case' })
    expect(withNoulCriteria({ true: 'yes', false: 'no' }, 'true', '')).toEqual({ false: 'no' })
    expect(withNoulCriteria({ true: 'yes' }, 'true', '')).toBeUndefined()
  })

  it('picks the next unused option name', () => {
    expect(nextChoiceName([])).toBe('option_1')
    expect(nextChoiceName([{ name: 'option_1', description: '' }])).toBe('option_2')
    expect(
      nextChoiceName([
        { name: 'option_1', description: '' },
        { name: 'option_2', description: '' },
      ]),
    ).toBe('option_3')
  })

  it('validates question readiness like the server does', () => {
    expect(isJevQuestionReady(undefined)).toBe(false)
    expect(isJevQuestionReady({ type: 'noul', instructions: '  ' })).toBe(false)
    expect(isJevQuestionReady({ type: 'noul', instructions: 'Is this intent?' })).toBe(true)
    expect(isJevQuestionReady({ type: 'choice', instructions: 'x', criteria: {} })).toBe(false)
    expect(
      isJevQuestionReady({ type: 'choice', instructions: 'x', criteria: { pro: 'y' } }),
    ).toBe(true)
    expect(isJevQuestionReady({ type: 'score', instructions: 'x', criteria: ['only'] })).toBe(false)
    expect(
      isJevQuestionReady({ type: 'score', instructions: 'x', criteria: ['low', 'high'] }),
    ).toBe(true)
  })
})
