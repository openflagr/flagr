import type { Flag, FlagView, Segment, Variant } from '@/api/types'

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

export function variantUsedInDistribution(flag: FlagView, variantId: number): boolean {
  return (flag.segments ?? []).some((s) =>
    (s.distributions ?? []).some((d) => d.variantID === variantId),
  )
}

function processVariant(variant: Variant): void {
  if (typeof variant.attachment === 'string') {
    try {
      variant.attachment = JSON.parse(variant.attachment) as Record<string, unknown>
    } catch {
      /* keep string */
    }
  }
}

export function normalizeSegment(segment: Segment): Segment {
  if (!segment.constraints) segment.constraints = []
  if (!segment.distributions) segment.distributions = []
  return segment
}

export function normalizeFlag(flag: Flag): FlagView {
  const f = { ...flag }
  if (!f.tags) f.tags = []
  if (!f.variants) f.variants = []
  if (!f.segments) f.segments = []
  f.variants.forEach((v) => processVariant(v))
  for (const segment of f.segments ?? []) {
    normalizeSegment(segment)
  }
  return f
}
