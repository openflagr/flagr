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
