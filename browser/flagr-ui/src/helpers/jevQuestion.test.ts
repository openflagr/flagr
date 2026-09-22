import { describe, expect, it } from 'vitest'
import {
  choiceCriteriaFromRows,
  choiceOperatorFor,
  choiceOptionsFromValue,
  choiceRowsFromCriteria,
  choiceValueFromOptions,
  defaultJevQuestion,
  formatJevMatch,
  formatJevSummary,
  isJevQuestionReady,
  isNegatedOperator,
  jevPropertyFor,
  jevPropertyName,
  nextChoiceName,
  noulCriteriaText,
  numberFromValue,
  operatorSymbol,
  reduceJevMatch,
  scoreLevelsFromCriteria,
  slugifyJevName,
  unquoteJevValue,
  withNoulCriteria,
  type JevMatchState,
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

  it('parses and formats choice constraint values', () => {
    expect(unquoteJevValue('"pro"')).toBe('pro')
    expect(choiceOptionsFromValue('"pro"')).toEqual(['pro'])
    expect(choiceOptionsFromValue('["pro","enterprise"]')).toEqual(['pro', 'enterprise'])
    expect(choiceOptionsFromValue('')).toEqual([])
    expect(choiceOptionsFromValue('not json')).toEqual(['not json'])
    expect(choiceValueFromOptions(['pro'])).toBe('"pro"')
    expect(choiceValueFromOptions(['pro', 'enterprise'])).toBe('["pro","enterprise"]')
    expect(choiceValueFromOptions([])).toBe('')
  })

  it('chooses choice operators from option count and negation', () => {
    expect(choiceOperatorFor(['pro'], false)).toBe('EQ')
    expect(choiceOperatorFor(['pro', 'enterprise'], false)).toBe('IN')
    expect(choiceOperatorFor(['pro'], true)).toBe('NEQ')
    expect(choiceOperatorFor(['pro', 'enterprise'], true)).toBe('NOTIN')
    expect(isNegatedOperator('NOTIN')).toBe(true)
    expect(isNegatedOperator('LT')).toBe(true)
    expect(isNegatedOperator('EQ')).toBe(false)
  })

  it('parses numeric values', () => {
    expect(numberFromValue('0.70')).toBeCloseTo(0.7)
    expect(numberFromValue('2')).toBe(2)
    expect(numberFromValue('nope')).toBeNull()
  })

  it('formats readable match and summary strings', () => {
    const noul = { type: 'noul' as const, instructions: 'x' }
    const choice = { type: 'choice' as const, instructions: 'x', criteria: { pro: 'y' } }
    const score = { type: 'score' as const, instructions: 'x', criteria: ['a', 'b'] }
    expect(operatorSymbol('GTE')).toBe('≥')
    expect(formatJevMatch(noul, 'GTE', '0.70')).toBe('P(true) ≥ 0.70')
    expect(formatJevMatch(choice, 'IN', '["pro"]')).toBe('in ["pro"]')
    expect(formatJevMatch(score, 'GTE', '1')).toBe('level ≥ 1')
    expect(formatJevSummary(noul, 'GTE', '0.70')).toBe('P(true) ≥ 0.70')
    expect(formatJevSummary(choice, 'EQ', '"pro"')).toBe('= "pro" · confidence ≥ 0.50')
    expect(
      formatJevSummary({ ...choice, confidenceThreshold: 0.8 }, 'EQ', '"pro"'),
    ).toBe('= "pro" · confidence ≥ 0.80')
  })
})

const noulState = (): JevMatchState => ({
  jev: { type: 'noul', instructions: 'x' },
  operator: 'GTE',
  value: '0.70',
})
const choiceState = (): JevMatchState => ({
  jev: { type: 'choice', instructions: 'x', criteria: { pro: 'y', free: 'z' } },
  operator: 'EQ',
  value: '"pro"',
})
const scoreState = (): JevMatchState => ({
  jev: { type: 'score', instructions: 'x', criteria: ['low', 'high'] },
  operator: 'GTE',
  value: '1',
})

describe('reduceJevMatch', () => {
  it('resets the match to type-appropriate defaults on type switch', () => {
    expect(reduceJevMatch(noulState(), { type: 'setType', questionType: 'choice' })).toMatchObject({
      operator: 'EQ',
      value: '',
    })
    expect(reduceJevMatch(noulState(), { type: 'setType', questionType: 'score' })).toMatchObject({
      operator: 'GTE',
      value: '0',
    })
    expect(reduceJevMatch(choiceState(), { type: 'setType', questionType: 'noul' })).toMatchObject({
      operator: 'GTE',
      value: '0.70',
    })
  })

  it('keeps instructions across a type switch', () => {
    expect(reduceJevMatch(noulState(), { type: 'setType', questionType: 'score' }).jev.instructions).toBe('x')
  })

  it('sets choice options and operator together (the clobbering bug)', () => {
    const one = reduceJevMatch(choiceState(), { type: 'setChoiceOptions', options: ['pro'] })
    expect(one.operator).toBe('EQ')
    expect(one.value).toBe('"pro"')

    const many = reduceJevMatch(choiceState(), { type: 'setChoiceOptions', options: ['pro', 'free'] })
    expect(many.operator).toBe('IN')
    expect(many.value).toBe('["pro","free"]')
  })

  it('toggles choice negation, keeping the options', () => {
    const neg = reduceJevMatch(choiceState(), { type: 'setChoiceNegate', negate: true })
    expect(neg.operator).toBe('NEQ')
    expect(neg.value).toBe('"pro"')
  })

  it('sets the scale level and preserves negation', () => {
    const atLeast = reduceJevMatch(scoreState(), { type: 'setScoreLevel', level: 0 })
    expect(atLeast.operator).toBe('GTE')
    expect(atLeast.value).toBe('0')

    const below = reduceJevMatch({ ...scoreState(), operator: 'LT' }, { type: 'setScoreLevel', level: 1 })
    expect(below.operator).toBe('LT')
    expect(below.value).toBe('1')

    expect(reduceJevMatch(scoreState(), { type: 'setScoreNegate', negate: true }).operator).toBe('LT')
  })

  it('applies confidence per type', () => {
    // noul: the slider is the P(true) threshold, stored in value
    expect(reduceJevMatch(noulState(), { type: 'setConfidence', value: 0.85 }).value).toBe('0.85')
    // choice/scale: the slider is the model confidence
    expect(
      reduceJevMatch(choiceState(), { type: 'setConfidence', value: 0.85 }).jev.confidenceThreshold,
    ).toBe(0.85)
    expect(
      reduceJevMatch(scoreState(), { type: 'setConfidence', value: 0.3 }).jev.confidenceThreshold,
    ).toBe(0.3)
  })

  it('sets the noul is / is not operator', () => {
    expect(reduceJevMatch(noulState(), { type: 'setNoulOperator', operator: 'LT' }).operator).toBe('LT')
  })

  it('updates instructions and criteria', () => {
    expect(reduceJevMatch(noulState(), { type: 'setInstructions', instructions: 'y' }).jev.instructions).toBe('y')
    expect(
      reduceJevMatch(noulState(), { type: 'setNoulCriteria', key: 'true', value: 'yes' }).jev.criteria,
    ).toEqual({ true: 'yes' })
    expect(
      reduceJevMatch(choiceState(), {
        type: 'setChoiceCriteria',
        rows: [{ name: 'a', description: 'b' }],
      }).jev.criteria,
    ).toEqual({ a: 'b' })
    expect(
      reduceJevMatch(scoreState(), { type: 'setScoreLevels', levels: ['a', 'b', 'c'] }).jev.criteria,
    ).toEqual(['a', 'b', 'c'])
  })
})
