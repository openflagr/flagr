import { findOperatorUi, type OperatorUiOption } from '@/helpers/constraintOperators'

export function propertyPlaceholderFor(
  operator: string,
  options: OperatorUiOption[],
): string {
  return findOperatorUi(operator, options)?.propertyPlaceholder ?? 'Property'
}

export function valuePlaceholderFor(operator: string, options: OperatorUiOption[]): string {
  return findOperatorUi(operator, options)?.valuePlaceholder ?? 'Value'
}