import constants from '@/helpers/constants'
import type { Flag, FlagView, Segment, Variant, VariantAttachment } from '@/api/types'

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

export function setEntityTypeOptionsFromApiKeys(
  keysFromApi: string[],
): { entityTypes: EntityTypeOption[]; allowCreateEntityType: boolean } {
  const pinned = constants.FLAGR_UI_POSSIBLE_ENTITY_TYPES
  if (pinned) {
    return {
      entityTypes: entityTypeOptionsFromKeys(pinned.split(',')),
      allowCreateEntityType: false,
    }
  }
  return {
    entityTypes: entityTypeOptionsFromKeys(keysFromApi),
    allowCreateEntityType: true,
  }
}

export function variantUsedInDistribution(flag: FlagView, variantId: number): boolean {
  return (flag.segments ?? []).some((s) =>
    (s.distributions ?? []).some((d) => d.variantID === variantId),
  )
}

function processVariant(variant: Variant): void {
  if (variant.attachmentValid === undefined) {
    variant.attachmentValid = true
  }
  if (typeof variant.attachment === 'string') {
    try {
      variant.attachment = JSON.parse(variant.attachment) as VariantAttachment
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
  for (const segment of f.segments) {
    normalizeSegment(segment)
  }
  return f as FlagView
}
