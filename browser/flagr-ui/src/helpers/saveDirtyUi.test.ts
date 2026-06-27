import { describe, expect, it } from 'vitest'
import { SAVE_DIRTY_TOOLTIP, saveButtonLabel, saveButtonType } from './saveDirtyUi'

describe('saveDirtyUi', () => {
  it('exposes reorder-aligned tooltip copy', () => {
    expect(SAVE_DIRTY_TOOLTIP).toContain('Save')
  })

  it('saveButtonLabel shows asterisk when dirty', () => {
    expect(saveButtonLabel(false)).toBe('Save')
    expect(saveButtonLabel(true)).toBe('Save *')
  })

  it('saveButtonType is warning only when dirty', () => {
    expect(saveButtonType(false)).toBeUndefined()
    expect(saveButtonType(true)).toBe('warning')
  })
})