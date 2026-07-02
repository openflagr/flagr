import { findOperatorUi, type OperatorUiOption } from '@/helpers/constraintOperators'
import { UI_STRING_CONTAINS, UI_STRING_NOT_CONTAINS } from '@/helpers/constraintOperatorSugar'

const API_BADGE: Record<string, string> = {
  EQ: '==',
  NEQ: '!=',
  LT: '<',
  LTE: '<=',
  GT: '>',
  GTE: '>=',
  IN: 'IN',
  NOTIN: 'NOT IN',
  CONTAINS: 'CONTAINS',
  NOTCONTAINS: 'NOT CONTAINS',
  EREG: '=~',
  NEREG: '!~',
  [UI_STRING_CONTAINS]: '=~',
  [UI_STRING_NOT_CONTAINS]: '!~',
}

/** API / expression token for el-tag in operator dropdown. */
export function operatorApiBadge(operatorValue: string): string {
  return API_BADGE[operatorValue] ?? operatorValue
}

/** Human label for closed select and dropdown text (no duplicated syntax in parens). */
export function operatorOptionDisplayText(op: OperatorUiOption): string {
  return op.label.replace(/\s*\([^)]*\)\s*$/, '').trim() || op.label
}

/** el-option :label — human text for a11y / search (no API token). */
export function operatorSelectLabel(op: OperatorUiOption): string {
  return operatorOptionDisplayText(op)
}

/** Closed el-select #label slot — API token for the chosen operator. */
export function operatorSelectClosedBadge(operatorValue: string | undefined): string {
  if (!operatorValue) return ''
  return operatorApiBadge(operatorValue)
}

/** One-line hint only when the operator is easy to misuse; otherwise omit. */
export function getOperatorHintLine(
  operatorValue: string | undefined,
  options?: OperatorUiOption[],
): string | null {
  const op = findOperatorUi(operatorValue, options)
  if (!op) return null

  switch (op.value) {
    case 'CONTAINS':
      return 'List/array property must contain this value — not substring search on a string.'
    case 'NOTCONTAINS':
      return 'List property must not contain this value.'
    case UI_STRING_CONTAINS:
      return 'Plain substring on a string; stored as =~ with escaped text.'
    case UI_STRING_NOT_CONTAINS:
      return 'Plain text must not appear in the string; stored as !~.'
    case 'EREG':
      return 'Regex on a string; for simple text prefer “text includes”.'
    case 'NEREG':
      return 'Regex must not match; for simple exclusion prefer “text excludes”.'
    case 'IN':
      return 'Value must be a JSON array of allowed scalars.'
    default:
      return null
  }
}