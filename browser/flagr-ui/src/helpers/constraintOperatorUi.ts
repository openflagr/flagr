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

/** One-line hint only when the operator is easy to misuse; otherwise omit. */
export function getOperatorHintLine(
  operatorValue: string | undefined,
  options?: OperatorUiOption[],
): string | null {
  const op = findOperatorUi(operatorValue, options)
  if (!op?.hintLine) return null
  return op.hintLine
}