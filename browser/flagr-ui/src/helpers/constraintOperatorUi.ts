import { findOperatorUi, type OperatorUiOption } from '@/helpers/constraintOperators'

/** API / expression token for el-tag in operator dropdown and closed select. */
export function operatorApiBadge(operatorValue: string, options?: OperatorUiOption[]): string {
  const op = findOperatorUi(operatorValue, options)
  return op?.exprToken ?? operatorValue
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
export function operatorSelectClosedBadge(
  operatorValue: string | undefined,
  options?: OperatorUiOption[],
): string {
  if (!operatorValue) return ''
  return operatorApiBadge(operatorValue, options)
}

/** Tooltip copy from catalog row (hintLine, else description). */
export function operatorHelpText(op: OperatorUiOption | undefined): string | null {
  if (!op) return null
  return op.hintLine ?? op.description ?? null
}

export function propertyPlaceholderFor(
  operator: string,
  options: OperatorUiOption[],
): string {
  return findOperatorUi(operator, options)?.propertyPlaceholder ?? 'Property'
}

export function valuePlaceholderFor(operator: string, options: OperatorUiOption[]): string {
  return findOperatorUi(operator, options)?.valuePlaceholder ?? 'Value'
}