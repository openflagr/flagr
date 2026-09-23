import { describe, expect, it } from 'vitest'
import type { JevQuestion } from '@/api/types'
import {
  DEFAULT_JEV_CONFIDENCE,
  DEFAULT_JEV_INSTRUCTIONS,
  defaultJevQuestion,
  jevQuestionProblems,
  noulCriteriaText,
  withNoulCriteria,
} from './jevQuestion'

const validNoul: JevQuestion = { type: 'noul', instructions: 'Is the plan enterprise?' }

describe('defaultJevQuestion', () => {
  it.each(['noul', 'choice', 'score'] as const)(
    'produces a question that passes validation for %s',
    (type) => {
      const question = defaultJevQuestion(type)
      expect(question.instructions).toBe(DEFAULT_JEV_INSTRUCTIONS[type])
      expect(String(question.instructions).trim()).not.toBe('')
      // A freshly added question must be saveable as-is, except for criteria the
      // user has to fill in (choice options / score levels).
      if (type === 'noul') {
        expect(jevQuestionProblems({ question })).toEqual([])
      }
      if (type === 'choice' || type === 'score') {
        expect(question.confidenceThreshold).toBe(DEFAULT_JEV_CONFIDENCE)
      }
    },
  )
})

describe('noul criteria', () => {
  it('round-trips true/false descriptions', () => {
    const criteria = withNoulCriteria(
      withNoulCriteria(undefined, 'true', 'about billing'),
      'false',
      'anything else',
    )
    expect(criteria).toEqual({ true: 'about billing', false: 'anything else' })
    expect(noulCriteriaText(criteria, 'true')).toBe('about billing')
    expect(noulCriteriaText(criteria, 'false')).toBe('anything else')
  })

  it('omits a blank description and returns undefined when both are blank', () => {
    expect(withNoulCriteria(undefined, 'true', '   ')).toBeUndefined()
    expect(withNoulCriteria(undefined, 'true', 'about billing')).toEqual({ true: 'about billing' })
  })
})

describe('jevQuestionProblems', () => {
  it('requires at least one question', () => {
    expect(jevQuestionProblems({})).toEqual(['Add at least one question.'])
  })

  it('accepts a well-formed noul question', () => {
    expect(jevQuestionProblems({ plan_tier: validNoul })).toEqual([])
  })

  it('rejects a name that is not identifier-like', () => {
    const problems = jevQuestionProblems({ 'plan-tier': validNoul })
    expect(problems).toHaveLength(1)
    expect(problems[0]).toContain('plan-tier')
  })

  it('rejects blank instructions', () => {
    const problems = jevQuestionProblems({ q: { type: 'noul', instructions: '   ' } })
    expect(problems).toEqual(['Question "q" needs instructions.'])
  })

  it('rejects a choice question with no options', () => {
    const problems = jevQuestionProblems({ q: { type: 'choice', instructions: 'Pick one.', criteria: {} } })
    expect(problems).toEqual(['Question "q" needs at least one option.'])
  })

  it('accepts a choice question with an option', () => {
    const question: JevQuestion = {
      type: 'choice',
      instructions: 'Pick one.',
      criteria: { pro: 'Pro plan' },
    }
    expect(jevQuestionProblems({ q: question })).toEqual([])
  })

  it('rejects a score question with fewer than two levels', () => {
    const problems = jevQuestionProblems({ q: { type: 'score', instructions: 'Rate.', criteria: ['low'] } })
    expect(problems).toEqual(['Question "q" needs between 2 and 10 levels.'])
  })

  it('accepts a score question with two levels', () => {
    const question: JevQuestion = {
      type: 'score',
      instructions: 'Rate.',
      criteria: ['low', 'high'],
    }
    expect(jevQuestionProblems({ q: question })).toEqual([])
  })

  it('rejects a confidence threshold outside [0,1]', () => {
    const question: JevQuestion = {
      type: 'choice',
      instructions: 'Pick one.',
      criteria: { pro: 'Pro' },
      confidenceThreshold: 2,
    }
    expect(jevQuestionProblems({ q: question })).toEqual(['Question "q" confidence must be between 0 and 1.'])
  })

  it('rejects a confidence threshold on a true/false question', () => {
    const problems = jevQuestionProblems({ q: { ...validNoul, confidenceThreshold: 0.5 } })
    expect(problems).toHaveLength(1)
    expect(problems[0]).toContain('already a 0-1 probability')
  })

  it('allows a confidence threshold on a choice question', () => {
    const question: JevQuestion = {
      type: 'choice',
      instructions: 'Pick one.',
      criteria: { pro: 'Pro' },
      confidenceThreshold: 0.5,
    }
    expect(jevQuestionProblems({ q: question })).toEqual([])
  })
})
