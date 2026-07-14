/** How long the copy control shows the success state before reverting. */
export const COPY_FEEDBACK_MS = 2000

/** Soft highlight duration when a deep-linked snapshot is scrolled into view. */
export const SNAPSHOT_HIGHLIGHT_MS = 1500

/**
 * Copy plain text to the clipboard.
 * Prefers `navigator.clipboard`; falls back to a temporary textarea + execCommand.
 * Returns whether the copy succeeded.
 */
export async function copyText(text: string): Promise<boolean> {
  if (typeof navigator !== 'undefined' && navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(text)
      return true
    } catch {
      // fall through to legacy path
    }
  }

  if (typeof document === 'undefined') {
    return false
  }

  try {
    const ta = document.createElement('textarea')
    ta.value = text
    ta.setAttribute('readonly', '')
    ta.style.position = 'fixed'
    ta.style.left = '-9999px'
    ta.style.top = '0'
    document.body.appendChild(ta)
    ta.focus()
    ta.select()
    const ok = document.execCommand('copy')
    document.body.removeChild(ta)
    return ok
  } catch {
    return false
  }
}
