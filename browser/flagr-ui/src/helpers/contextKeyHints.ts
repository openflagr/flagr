const DAYS = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']
const MONTHS = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
]

/**
 * Format a Unix-epoch-second integer as a UTC datetime string.
 *   e.g. 1752556800 → "Jul 15, 2026 00:00:00 UTC"
 */
function formatTsEpoch(value: string): string | null {
  const n = Number(value)
  if (!Number.isInteger(n)) return null
  const d = new Date(n * 1000)
  if (isNaN(d.getTime())) return null
  return d.toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
    timeZone: 'UTC',
    timeZoneName: 'short',
  })
}

/**
 * Return a human-readable hint for a built-in context key, or null.
 */
export function contextKeyHint(property: string, value: string): string | null {
  if (!property || !value) return null

  if (property === '@ts') {
    const formatted = formatTsEpoch(value)
    return formatted ? `= ${formatted}` : null
  }

  if (property === '@ts_hour') {
    const n = Number(value)
    if (!Number.isInteger(n) || n < 0 || n > 23) return null
    const h12 = n === 0 ? 12 : n > 12 ? n - 12 : n
    const ampm = n < 12 ? 'AM' : 'PM'
    return `= ${String(n).padStart(2, '0')}:00 UTC (${h12} ${ampm})`
  }

  if (property === '@ts_weekday') {
    const n = Number(value)
    if (!Number.isInteger(n) || n < 0 || n > 6) return null
    return `= ${DAYS[n]}`
  }

  if (property === '@ts_month') {
    const n = Number(value)
    if (!Number.isInteger(n) || n < 1 || n > 12) return null
    return `= ${MONTHS[n - 1]}`
  }

  return null
}
