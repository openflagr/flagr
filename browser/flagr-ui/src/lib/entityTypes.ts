export interface EntityTypeOption {
  label: string
  value: string
}

export function entityTypeOptionsFromKeys(keys: string[]): EntityTypeOption[] {
  const arr = keys.map((key) => ({
    label: key === '' ? '<null>' : key,
    value: key,
  }))
  if (!keys.includes('')) {
    arr.unshift({ label: '<null>', value: '' })
  }
  return arr
}