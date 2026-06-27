/** Matches segment Reorder: subtle hint that local edits are not persisted yet. */
export const SAVE_DIRTY_TOOLTIP = 'Unsaved changes — click Save to persist'

export function saveButtonLabel(dirty: boolean): string {
  return dirty ? 'Save *' : 'Save'
}

export function saveButtonType(dirty: boolean): 'warning' | undefined {
  return dirty ? 'warning' : undefined
}