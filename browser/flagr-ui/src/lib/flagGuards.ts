import type { FlagView, Variant } from '@/types'

export function variantUsedInDistribution(flag: FlagView, variantId: number): boolean {
  return (flag.segments ?? []).some((s) =>
    (s.distributions ?? []).some((d) => d.variantID === variantId),
  )
}