import type { Constraint } from '@/api/types'
import type { OperatorValue } from '@/api/types'

/** UI-only operator ids; persisted as EREG / NEREG with a literal regex value. */
export const UI_STRING_CONTAINS = 'UI_STRING_CONTAINS' as const
export const UI_STRING_NOT_CONTAINS = 'UI_STRING_NOT_CONTAINS' as const

export type UiSugarOperator = typeof UI_STRING_CONTAINS | typeof UI_STRING_NOT_CONTAINS

export type ConstraintUiOperator = OperatorValue | UiSugarOperator

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
  const inner = stripJsonStringQuotes(storedValue)
  if (inner == null) return false
  if (/[*+?|()[\]^$]/.test(inner)) return false
  if (/\.\*|\.\+/.test(inner)) return false
  return true
}

/** Plain text for the value input when editing a literal substring constraint. */
export function plainTextFromStoredConstraintValue(storedValue: string): string {
  const inner = stripJsonStringQuotes(storedValue.trim())
  if (inner != null && isStoredLiteralSubstringPattern(storedValue)) return inner
  return storedValue
}

/** Which operator to show in the UI select for a persisted constraint. */
export function resolveUiOperator(operator: string, value: string): ConstraintUiOperator {
  if (operator === 'EREG' && isStoredLiteralSubstringPattern(value)) return UI_STRING_CONTAINS
  if (operator === 'NEREG' && isStoredLiteralSubstringPattern(value)) return UI_STRING_NOT_CONTAINS
  return operator as OperatorValue
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

  if (constraint.operator === UI_STRING_CONTAINS) {
    const plain = isPlainDraftValue(rawValue)
      ? rawValue.trim()
      : plainTextFromStoredConstraintValue(rawValue)
    return {
      ...constraint,
      property,
      operator: 'EREG',
      value: `"${escapeRegexLiteral(plain)}"`,
    }
  }
  if (constraint.operator === UI_STRING_NOT_CONTAINS) {
    const plain = isPlainDraftValue(rawValue)
      ? rawValue.trim()
      : plainTextFromStoredConstraintValue(rawValue)
    return {
      ...constraint,
      property,
      operator: 'NEREG',
      value: `"${escapeRegexLiteral(plain)}"`,
    }
  }

  return { ...constraint, property, value: rawValue }
}

/** Apply UI operator selection to the in-memory constraint (sugar ids stay until save). */
export function applyUiOperatorSelection(constraint: Constraint, uiOperator: string): void {
  if (uiOperator === UI_STRING_CONTAINS || uiOperator === UI_STRING_NOT_CONTAINS) {
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