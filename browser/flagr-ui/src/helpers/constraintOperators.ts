import operatorsData from '@/operators.json'
import type { OperatorValue } from '@/api/types'
import {
  UI_STRING_CONTAINS,
  UI_STRING_NOT_CONTAINS,
  type UiSugarOperator,
} from '@/helpers/constraintOperatorSugar'

/** Catalog row from operators.json (API + optional UI-only sugar). */
export interface OperatorCatalogRow {
  value: string
  label: string
  group: string
  description: string
  exprToken?: string
  hintLine?: string
  propertyPlaceholder?: string
  valuePlaceholder?: string
  uiOnly?: boolean
  persistAs?: OperatorValue
}

export type OperatorUiValue = OperatorValue | UiSugarOperator

/** UI metadata for a constraint operator option in el-select. */
export interface OperatorUiOption {
  value: OperatorUiValue
  label: string
  group: string
  description: string
  exprToken: string
  hintLine?: string
  propertyPlaceholder?: string
  valuePlaceholder?: string
  uiOnly?: boolean
  persistAs?: OperatorValue
}

const GROUP_ORDER = ['Compare', 'Lists', 'Text (simple)', 'Text pattern'] as const

function rowToUiOption(row: OperatorCatalogRow): OperatorUiOption {
  return {
    value: row.value as OperatorUiValue,
    label: row.label,
    group: row.group,
    description: row.description,
    exprToken: row.exprToken ?? row.value,
    hintLine: row.hintLine,
    propertyPlaceholder: row.propertyPlaceholder,
    valuePlaceholder: row.valuePlaceholder,
    uiOnly: row.uiOnly,
    persistAs: row.persistAs,
  }
}

const CATALOG_ROWS = operatorsData.operators as OperatorCatalogRow[]

export const OPERATOR_UI_OPTIONS: OperatorUiOption[] = CATALOG_ROWS.map(rowToUiOption)

export function findOperatorUi(
  value: string | undefined,
  options: OperatorUiOption[] = OPERATOR_UI_OPTIONS,
): OperatorUiOption | undefined {
  if (!value) return undefined
  return options.find((o) => o.value === value)
}

export interface OperatorOptionGroup {
  label: string
  options: OperatorUiOption[]
}

/** Groups operators for el-select option groups; preserves GROUP_ORDER. */
export function operatorOptionGroups(
  options: OperatorUiOption[] = OPERATOR_UI_OPTIONS,
): OperatorOptionGroup[] {
  const byGroup: Record<string, OperatorUiOption[]> = {}
  for (const op of options) {
    if (!byGroup[op.group]) byGroup[op.group] = []
    byGroup[op.group].push(op)
  }
  const groups: OperatorOptionGroup[] = []
  for (const label of GROUP_ORDER) {
    const opts = byGroup[label]
    if (opts?.length) groups.push({ label, options: opts })
  }
  for (const label of Object.keys(byGroup)) {
    if (!(GROUP_ORDER as readonly string[]).includes(label)) {
      groups.push({ label, options: byGroup[label] })
    }
  }
  return groups
}

export { UI_STRING_CONTAINS, UI_STRING_NOT_CONTAINS }