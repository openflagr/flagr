import operatorsData from '@/operators.json'
import type { OperatorValue } from '@/api/types'
import {
  UI_STRING_CONTAINS,
  UI_STRING_NOT_CONTAINS,
} from '@/helpers/constraintOperatorSugar'

/** UI metadata for a constraint operator; API values unchanged; sugar ids map to EREG/NEREG on save. */
export interface OperatorUiOption {
  value: string
  label: string
  group: string
  description: string
  propertyPlaceholder?: string
  valuePlaceholder?: string
}

const GROUP_ORDER = ['Compare', 'Lists', 'Text (simple)', 'Text pattern'] as const

const API_OPERATOR_UI_OPTIONS: OperatorUiOption[] = operatorsData.operators as OperatorUiOption[]

const SUGAR_OPERATOR_UI_OPTIONS: OperatorUiOption[] = [
  {
    value: UI_STRING_CONTAINS,
    label: 'Text includes',
    group: 'Text (simple)',
    description: 'String must contain this plain text (saved as =~ with escaped text).',
    propertyPlaceholder: 'email',
    valuePlaceholder: '@gmail.com',
  },
  {
    value: UI_STRING_NOT_CONTAINS,
    label: 'Text excludes',
    group: 'Text (simple)',
    description: 'String must not contain this plain text (saved as !~).',
    propertyPlaceholder: 'user_agent',
    valuePlaceholder: 'bot',
  },
]

export const OPERATOR_UI_OPTIONS: OperatorUiOption[] = [
  ...API_OPERATOR_UI_OPTIONS,
  ...SUGAR_OPERATOR_UI_OPTIONS,
]

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

export function isApiOperatorValue(value: string): value is OperatorValue {
  return API_OPERATOR_UI_OPTIONS.some((o) => o.value === value)
}