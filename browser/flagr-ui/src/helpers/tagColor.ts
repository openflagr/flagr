/**
 * Deterministic colour for a tag value.  Hashes the string to a hue (0-360)
 * and returns an HSL background colour suitable for el-tag :style.
 */
export function tagColor(value: string): string {
  let h = 0
  for (let i = 0; i < value.length; i++) {
    h = value.charCodeAt(i) + ((h << 5) - h)
  }
  const hue = ((h % 360) + 360) % 360
  return `hsl(${hue}, 55%, 92%)`
}
