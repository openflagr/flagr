import type { Constraint } from '@/api/types'
import type { OperatorValue } from '@/api/types'

/** UI-only operator ids; persisted as EREG / NEREG with a literal regex value. */
export const UI_STRING_CONTAINS = 'UI_STRING_CONTAINS' as const
export const UI_STRING_NOT_CONTAINS = 'UI_STRING_NOT_CONTAINS' as const

export type UiSugarOperator = typeof UI_STRING_CONTAINS | typeof UI_STRING_NOT_CONTAINS

export type ConstraintUiOperator = OperatorValue | UiSugarOperator
export function isUiSugarOperator(value: string): value is UiSugarOperator {
  return value === UI_STRING_CONTAINS || value === UI_STRING_NOT_CONTAINS
}
const REGEX_SPECIAL = /[.*+?^${}()|[\]\\]/g

/** Escape plain text for use as a Go/JS regexp literal substring pattern. */
export function escapeRegexLiteral(plain: string): string {
  return plain.replace(REGEX_SPECIAL, '\\$&')
}

function stripJsonStringQuotes(raw: string): string | null {
  const t = raw.trim()
  if (t.length < 2 || t[0] !== '"' || t[t.length - 1] !== '"') return null
  try {
    return JSON.parse(t) as string
  } catch {
    return t.slice(1, -1)
  }
}

/** True when stored EREG/NEREG value is a quoted literal suitable for the string-contains sugar. */
export function isStoredLiteralSubstringPattern(storedValue: string): boolean {
  const inner = stripJsonStringQuotes(storedValue.trim())
  if (inner == null) return false
  return escapeRegexLiteral(inner) === inner
}

/** Plain text for the value input when editing a literal substring constraint. */
export function plainTextFromStoredConstraintValue(storedValue: string): string {
  const inner = stripJsonStringQuotes(storedValue.trim())
  return inner ?? storedValue.trim()
}

/** Which operator to show in the UI select for a persisted constraint. */
export function resolveUiOperator(operator: string, value: string): ConstraintUiOperator {
  if (operator === 'EREG' && isStoredLiteralSubstringPattern(value)) {
    return UI_STRING_CONTAINS
  }
  if (operator === 'NEREG' && isStoredLiteralSubstringPattern(value)) {
    return UI_STRING_NOT_CONTAINS
  }
  return operator as ConstraintUiOperator
}

/** Value shown in the constraint value field (plain text for sugar / literal EREG). */
export function constraintValueForInput(constraint: Constraint): string {
  const ui = resolveUiOperator(constraint.operator, constraint.value)
  if (ui === UI_STRING_CONTAINS || ui === UI_STRING_NOT_CONTAINS) {
    return plainTextFromStoredConstraintValue(constraint.value)
  }
  return constraint.value
}

function isPlainDraftValue(value: string): boolean {
  const t = value.trim()
  if (!t) return true
  return stripJsonStringQuotes(t) == null
}

/** Normalize constraint immediately before API create/update. */
export function materializeConstraintForApi(constraint: Constraint): Constraint {
  const property = constraint.property.trim()
  const rawValue = constraint.value.trim()

  if (isUiSugarOperator(constraint.operator)) {
    const plain = isPlainDraftValue(rawValue)
      ? rawValue.trim()
      : plainTextFromStoredConstraintValue(rawValue)
    const apiOp: OperatorValue =
      constraint.operator === UI_STRING_CONTAINS ? 'EREG' : 'NEREG'
    return {
      ...constraint,
      property,
      operator: apiOp,
      value: `"${escapeRegexLiteral(plain)}"`,
    }
  }

  return { ...constraint, property, value: rawValue }
}

/**
 * Apply UI operator selection on the in-memory constraint (parent-owned object).
 * Sugar ids remain on constraint.operator until save; materializeConstraintForApi runs on persist.
 */
export function applyUiOperatorSelection(constraint: Constraint, uiOperator: string): void {
  if (isUiSugarOperator(uiOperator)) {
    if (
      (constraint.operator === 'EREG' || constraint.operator === 'NEREG') &&
      isStoredLiteralSubstringPattern(constraint.value)
    ) {
      constraint.value = plainTextFromStoredConstraintValue(constraint.value)
    }
    constraint.operator = uiOperator
    return
  }
  constraint.operator = uiOperator
}