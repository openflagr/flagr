import type { Flag, FlagView, Segment, Variant } from '@/types'

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