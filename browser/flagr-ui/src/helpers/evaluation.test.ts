import { describe, expect, it } from 'vitest'
import { editorTextFromValue, toJsonText } from './evaluation'

describe('toJsonText', () => {
  it('pretty-prints an object with two-space indentation', () => {
    expect(toJsonText({ a: 1, b: { c: true } })).toBe(
      '{\n  "a": 1,\n  "b": {\n    "c": true\n  }\n}',
    )
  })

  it('falls back to an empty string for values JSON.stringify drops', () => {
    expect(toJsonText(undefined)).toBe('')
  })
})

describe('editorTextFromValue', () => {
  it('returns null when the value is the echo of the editor text', () => {
    // Reformatting on this echo would move the caret.
    const value = { entityID: 'a1234', enableDebug: true }
    expect(editorTextFromValue(value, toJsonText(value))).toBeNull()
  })

  it('returns the text when the value actually changed', () => {
    expect(editorTextFromValue({ entityID: 'b5678' }, toJsonText({ entityID: 'a1234' }))).toBe(
      '{\n  "entityID": "b5678"\n}',
    )
  })
})
