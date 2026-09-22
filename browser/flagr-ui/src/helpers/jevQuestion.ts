import type { JevQuestion, JevQuestionType } from '@/api/types'

/** entityContext property prefix for Jev answers: question `foo` → `@jev.foo`. */
export const JEV_PROPERTY_PREFIX = '@jev.'

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
  if (type === 'choice') return { type, instructions: '', criteria: {} }
  if (type === 'score') return { type, instructions: '', criteria: ['', ''] }
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

/** Default confidence gate when a question does not set one. */
export const DEFAULT_JEV_CONFIDENCE = 0.5

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

/** Readable summary of the answer match, e.g. `P(yes) ≥ 0.70` or `in ["pro"]`. */
export function formatJevMatch(
  jev: JevQuestion | undefined,
  operator: string,
  value: string,
): string {
  const symbol = operatorSymbol(operator)
  if (jev?.type === 'noul') return `P(yes) ${symbol} ${value}`
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
