import type { Enricher } from '@/api/types'

/** A namespace's enriched properties, for the constraint property picker. */
export interface EnricherPropertyGroup {
  namespace: string
  label: string
  options: string[]
}

/**
 * Groups the enriched properties of a flag's effective enricher catalog by
 * namespace, so the constraint picker can offer `@jev_plan_tier`,
 * `@ts_hour`, `@http_*`, ... without the user reading the docs.
 */
export function enricherPropertyGroups(enrichers: Enricher[] | undefined): EnricherPropertyGroup[] {
  return (enrichers ?? [])
    .filter((enricher) => (enricher.properties ?? []).length > 0)
    .map((enricher) => ({
      namespace: enricher.namespace,
      label: enricher.namespace,
      options: [...(enricher.properties ?? [])].sort(),
    }))
}

/** Flat list of all enriched property names, in catalog order. */
export function enrichedProperties(enrichers: Enricher[] | undefined): string[] {
  return enricherPropertyGroups(enrichers).flatMap((group) => group.options)
}
