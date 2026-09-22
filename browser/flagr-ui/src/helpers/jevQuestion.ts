import type { JevQuestion, JevQuestionType } from '@/api/types'

/** entityContext property prefix for Jev answers: question `foo` → `@jev.foo`. */
export const JEV_PROPERTY_PREFIX = '@jev.'

/** Default confidence gate for a choice/score question. */
export const DEFAULT_JEV_CONFIDENCE = 0.5

export interface ChoiceRow {
  name: string
  description: string
}

/** Strip the `@jev.` prefix from a constraint property. */
export function jevPropertyName(property: string): string {
  return property.startsWith(JEV_PROPERTY_PREFIX)
    ? property.slice(JEV_PROPERTY_PREFIX.length)
    : ''
}

/** Build the `@jev.<slug>` property for a question name. */
export function jevPropertyFor(name: string): string {
  return JEV_PROPERTY_PREFIX + slugifyJevName(name)
}

/** Restrict a question name to the safe character set. */
export function slugifyJevName(name: string): string {
  return name.replace(/[^a-zA-Z0-9_]/g, '_')
}

/** A fresh question for the given type, with the criteria shape that type needs. */
export function defaultJevQuestion(type: JevQuestionType): JevQuestion {
  if (type === 'choice') {
    return { type, instructions: '', criteria: {}, confidenceThreshold: DEFAULT_JEV_CONFIDENCE }
  }
  if (type === 'score') {
    return { type, instructions: '', criteria: ['', ''], confidenceThreshold: DEFAULT_JEV_CONFIDENCE }
  }
  return { type, instructions: '' }
}

/** Choice criteria (`{option: description}`) → editable rows. */
export function choiceRowsFromCriteria(criteria: unknown): ChoiceRow[] {
  if (!criteria || typeof criteria !== 'object' || Array.isArray(criteria)) return []
  return Object.entries(criteria as Record<string, unknown>).map(([name, description]) => ({
    name,
    description: typeof description === 'string' ? description : JSON.stringify(description),
  }))
}

/** Editable rows → choice criteria. */
export function choiceCriteriaFromRows(rows: ChoiceRow[]): Record<string, string> {
  const criteria: Record<string, string> = {}
  for (const row of rows) criteria[row.name] = row.description
  return criteria
}

/** Score criteria (ordered level array) → editable level strings. */
export function scoreLevelsFromCriteria(criteria: unknown): string[] {
  if (!Array.isArray(criteria)) return []
  return criteria.map((level) => (typeof level === 'string' ? level : JSON.stringify(level)))
}

/** Read a Noul `true`/`false` description. */
export function noulCriteriaText(criteria: unknown, key: 'true' | 'false'): string {
  if (!criteria || typeof criteria !== 'object' || Array.isArray(criteria)) return ''
  const value = (criteria as Record<string, unknown>)[key]
  return typeof value === 'string' ? value : ''
}

/** Set/clear a Noul `true`/`false` description, returning undefined when empty. */
export function withNoulCriteria(
  criteria: unknown,
  key: 'true' | 'false',
  value: string,
): Record<string, unknown> | undefined {
  const base: Record<string, unknown> = {}
  if (criteria && typeof criteria === 'object' && !Array.isArray(criteria)) {
    Object.assign(base, criteria)
  }
  if (value.trim() === '') delete base[key]
  else base[key] = value
  return Object.keys(base).length > 0 ? base : undefined
}

/** Next unused default option name (`option_N`). */
export function nextChoiceName(rows: ChoiceRow[]): string {
  let n = rows.length + 1
  while (rows.some((row) => row.name === `option_${n}`)) n += 1
  return `option_${n}`
}

/**
 * Client-side readiness check mirroring the server validation: a question needs
 * instructions and, for choice/score, non-empty criteria.
 */
export function isJevQuestionReady(jev: JevQuestion | undefined): boolean {
  if (!jev) return false
  if (typeof jev.instructions !== 'string' || jev.instructions.trim() === '') return false
  if (jev.type === 'choice') {
    const criteria = jev.criteria
    return (
      !!criteria &&
      typeof criteria === 'object' &&
      !Array.isArray(criteria) &&
      Object.keys(criteria).length > 0
    )
  }
  if (jev.type === 'score') {
    return Array.isArray(jev.criteria) && jev.criteria.length >= 2
  }
  return true
}

/** Strip surrounding JSON string quotes from a constraint value. */
export function unquoteJevValue(value: string): string {
  const trimmed = value.trim()
  if (trimmed.length >= 2 && trimmed.startsWith('"') && trimmed.endsWith('"')) {
    try {
      return JSON.parse(trimmed) as string
    } catch {
      return trimmed.slice(1, -1)
    }
  }
  return trimmed
}

/** Parse a choice constraint value into option names (single or JSON array). */
export function choiceOptionsFromValue(value: string): string[] {
  const trimmed = value.trim()
  if (!trimmed) return []
  if (trimmed.startsWith('[')) {
    try {
      const parsed: unknown = JSON.parse(trimmed)
      if (Array.isArray(parsed)) return parsed.map((option) => String(option))
    } catch {
      return []
    }
  }
  const single = unquoteJevValue(trimmed)
  return single ? [single] : []
}

/** Build a choice constraint value from option names: `"a"` or `["a","b"]`. */
export function choiceValueFromOptions(options: string[]): string {
  if (options.length === 0) return ''
  if (options.length === 1) return JSON.stringify(options[0])
  return JSON.stringify(options)
}

/** Operator for a choice match: EQ/NEQ for one option, IN/NOTIN for several. */
export function choiceOperatorFor(options: string[], negate: boolean): string {
  if (negate) return options.length > 1 ? 'NOTIN' : 'NEQ'
  return options.length > 1 ? 'IN' : 'EQ'
}

/** Human-readable symbol for a constraint operator. */
export function operatorSymbol(operator: string): string {
  switch (operator) {
    case 'GTE':
      return '≥'
    case 'GT':
      return '>'
    case 'LTE':
      return '≤'
    case 'LT':
      return '<'
    case 'NEQ':
      return '≠'
    case 'NOTIN':
      return 'not in'
    case 'IN':
      return 'in'
    case 'EQ':
      return '='
    default:
      return operator
  }
}

/** Readable summary of the answer match, e.g. `P(true) ≥ 0.70` or `in ["pro"]`. */
export function formatJevMatch(
  jev: JevQuestion | undefined,
  operator: string,
  value: string,
): string {
  const symbol = operatorSymbol(operator)
  if (jev?.type === 'noul') return `P(true) ${symbol} ${value}`
  if (jev?.type === 'score') return `level ${symbol} ${value}`
  return `${symbol} ${value}`
}

/**
 * One-line summary for the constraint row. Noul has no separate confidence, so
 * its probability threshold is the whole story; choice/score append the gate.
 */
export function formatJevSummary(
  jev: JevQuestion | undefined,
  operator: string,
  value: string,
): string {
  const match = formatJevMatch(jev, operator, value)
  if (!jev || jev.type === 'noul') return match
  const confidence = jev.confidenceThreshold ?? DEFAULT_JEV_CONFIDENCE
  return `${match} · confidence ≥ ${confidence.toFixed(2)}`
}

/** Parse a numeric constraint value (noul threshold, score level). */
export function numberFromValue(value: string): number | null {
  const parsed = Number(value.trim())
  return Number.isFinite(parsed) ? parsed : null
}

/** Operators that negate a choice/score match. */
export function isNegatedOperator(operator: string): boolean {
  return operator === 'NEQ' || operator === 'NOTIN' || operator === 'LT' || operator === 'LTE'
}

/** Direction of a choice match as shown in the UI: any of / none of. */
export type ChoiceDirection = 'include' | 'exclude'

/** Direction of a scale match as shown in the UI: at least / below. */
export type ScaleDirection = 'atleast' | 'below'

/** Map a choice operator to its UI direction (any of = include, none of = exclude). */
export function choiceDirectionFromOperator(operator: string): ChoiceDirection {
  return isNegatedOperator(operator) ? 'exclude' : 'include'
}

/** Map the UI choice direction back to the negate flag. */
export function choiceNegateFromDirection(direction: string): boolean {
  return direction === 'exclude'
}

/** Map a scale operator to its UI direction (≥ = at least, < = below). */
export function scaleDirectionFromOperator(operator: string): ScaleDirection {
  return operator === 'LT' || operator === 'LTE' ? 'below' : 'atleast'
}

/** Map the UI scale direction back to the negate flag. */
export function scaleNegateFromDirection(direction: string): boolean {
  return direction === 'below'
}

/** Full editable state of a Jev constraint: the question plus its match. */
export interface JevMatchState {
  jev: JevQuestion
  operator: string
  value: string
}

/** Actions the Jev editor dispatches. Kept pure so every question type is testable. */
export type JevAction =
  | { type: 'setType'; questionType: JevQuestionType }
  | { type: 'setInstructions'; instructions: string }
  | { type: 'setNoulCriteria'; key: 'true' | 'false'; value: string }
  | { type: 'setChoiceCriteria'; rows: ChoiceRow[] }
  | { type: 'setScoreLevels'; levels: string[] }
  | { type: 'setChoiceOptions'; options: string[] }
  | { type: 'setChoiceNegate'; negate: boolean }
  | { type: 'setScoreLevel'; level: number }
  | { type: 'setScoreNegate'; negate: boolean }
  | { type: 'setNoulOperator'; operator: string }
  | { type: 'setConfidence'; value: number }
  | { type: 'setJev'; jev: JevQuestion }

const DEFAULT_NOUL_THRESHOLD = 0.7

/** Default match (operator + value) for a question type. */
function defaultMatchFor(type: JevQuestionType): { operator: string; value: string } {
  if (type === 'choice') return { operator: 'EQ', value: '' }
  if (type === 'score') return { operator: 'GTE', value: '0' }
  return { operator: 'GTE', value: DEFAULT_NOUL_THRESHOLD.toFixed(2) }
}

/**
 * Patch applied when the JEV toggle flips on: seed a question and a
 * type-compatible match. Shared by the add and existing constraint rows.
 */
export function jevEnablePatch(
  current: JevQuestion | undefined,
  property: string,
): { jev: JevQuestion; property: string; operator: string; value: string } {
  return {
    jev: current ?? defaultJevQuestion('noul'),
    property: jevPropertyFor(jevPropertyName(property) || 'question'),
    ...defaultMatchFor('noul'),
  }
}

/**
 * Pure state transition for the Jev editor. Emitting operator and value in one
 * step is what keeps choice/scale edits from clobbering each other.
 */
export function reduceJevMatch(state: JevMatchState, action: JevAction): JevMatchState {
  switch (action.type) {
    case 'setType': {
      const jev = {
        ...defaultJevQuestion(action.questionType),
        instructions: state.jev.instructions,
      }
      return { jev, ...defaultMatchFor(action.questionType) }
    }
    case 'setInstructions':
      return { ...state, jev: { ...state.jev, instructions: action.instructions } }
    case 'setNoulCriteria':
      return {
        ...state,
        jev: {
          ...state.jev,
          criteria: withNoulCriteria(state.jev.criteria, action.key, action.value),
        },
      }
    case 'setChoiceCriteria':
      return { ...state, jev: { ...state.jev, criteria: choiceCriteriaFromRows(action.rows) } }
    case 'setScoreLevels':
      return { ...state, jev: { ...state.jev, criteria: action.levels } }
    case 'setChoiceOptions':
      return {
        ...state,
        value: choiceValueFromOptions(action.options),
        operator: choiceOperatorFor(action.options, isNegatedOperator(state.operator)),
      }
    case 'setChoiceNegate':
      return {
        ...state,
        operator: choiceOperatorFor(choiceOptionsFromValue(state.value), action.negate),
      }
    case 'setScoreLevel':
      return {
        ...state,
        value: String(action.level),
        operator: isNegatedOperator(state.operator) ? 'LT' : 'GTE',
      }
    case 'setScoreNegate':
      return { ...state, operator: action.negate ? 'LT' : 'GTE' }
    case 'setNoulOperator':
      return { ...state, operator: action.operator }
    case 'setConfidence':
      if (state.jev.type === 'noul') return { ...state, value: action.value.toFixed(2) }
      return { ...state, jev: { ...state.jev, confidenceThreshold: action.value } }
    case 'setJev':
      // Keep the authored question; reset the match only when the type changed,
      // so a pasted question can't keep an incompatible operator/value.
      if (action.jev.type === state.jev.type) return { ...state, jev: action.jev }
      return { ...state, jev: action.jev, ...defaultMatchFor(action.jev.type) }
  }
}
