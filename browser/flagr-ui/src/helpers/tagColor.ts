/**
 * Deterministic colour for a tag value. Hashes the string to a hue (0–360)
 * and returns a soft HSL background suitable for el-tag :style.
 *
 * Saturation/lightness stay muted so tags read as chips, not neon badges
 * (easier on the eye against the soft light theme).
 */
const TAG_SATURATION_PCT = 38
const TAG_LIGHTNESS_PCT = 90

export function tagColor(value: string): string {
  let h = 0
  for (let i = 0; i < value.length; i++) {
    h = value.charCodeAt(i) + ((h << 5) - h)
  }
  const hue = ((h % 360) + 360) % 360
  return `hsl(${hue}, ${TAG_SATURATION_PCT}%, ${TAG_LIGHTNESS_PCT}%)`
}
