/**
 * Read-only data source for eval-only mode (json_file / json_http).
 *
 * In that mode the CRUD API is not registered — the only fully populated
 * data plane is `GET /export/eval_cache/json`, which serves the live
 * EvalCache in the GitOps `entity.Flag` shape (PascalCase, `{ Flags: [...] }`).
 * This module fetches that dump once, maps it to the swagger camelCase types
 * the rest of the UI speaks, and answers the read calls in `crud.ts`.
 * The export JSON shape itself is the GitOps source of truth and must not be
 * changed to suit the UI — all adaptation happens here.
 */
import type { Flag, Segment, Tag, Variant } from './types'
import type { ApiResult } from './result'
import { ok, err } from './result'
import { ApiHttpError } from './errors'
import { requestJson } from './http'

/** `pkg/entity` structs as serialized by the export endpoint (Go field names). */
interface ExportedTag {
  ID: number
  Value: string
}

interface ExportedConstraint {
  ID: number
  Property: string
  Operator: string
  Value: string
}

interface ExportedDistribution {
  ID: number
  VariantID: number
  VariantKey: string
  Percent: number
}

interface ExportedSegment {
  ID: number
  Description: string
  Rank: number
  RolloutPercent: number
  Constraints?: ExportedConstraint[] | null
  Distributions?: ExportedDistribution[] | null
}

interface ExportedVariant {
  ID: number
  Key: string
  Attachment?: Record<string, unknown> | null
}

interface ExportedFlag {
  ID: number
  Key: string
  Description: string
  CreatedBy?: string
  UpdatedBy?: string
  UpdatedAt?: string
  Enabled: boolean
  Notes?: string
  DataRecordsEnabled?: boolean
  EntityType?: string
  Tags?: ExportedTag[] | null
  Variants?: ExportedVariant[] | null
  Segments?: ExportedSegment[] | null
}

interface EvalCacheDump {
  Flags?: ExportedFlag[] | null
}

function mapSegment(s: ExportedSegment): Segment {
  return {
    id: s.ID,
    description: s.Description,
    rank: s.Rank,
    rolloutPercent: s.RolloutPercent,
    constraints: (s.Constraints ?? []).map((c) => ({
      id: c.ID,
      property: c.Property,
      operator: c.Operator,
      value: c.Value,
    })),
    distributions: (s.Distributions ?? []).map((d) => ({
      id: d.ID,
      percent: d.Percent,
      variantID: d.VariantID,
      variantKey: d.VariantKey,
    })),
  }
}

function mapVariant(v: ExportedVariant): Variant {
  return {
    id: v.ID,
    key: v.Key,
    attachment: v.Attachment ?? undefined,
  }
}

/**
 * Map one exported flag to the swagger-shaped `Flag`. GORM-only fields
 * (DeletedAt, SnapshotID, FlagID, SegmentID, CreatedAt) are ignored.
 * Segment order is kept as exported: in json mode the source order is the
 * evaluation order, so the UI must show it unchanged.
 */
export function mapExportedFlag(e: ExportedFlag): Flag {
  return {
    id: e.ID,
    key: e.Key,
    description: e.Description,
    enabled: e.Enabled,
    notes: e.Notes,
    createdBy: e.CreatedBy,
    updatedBy: e.UpdatedBy,
    updatedAt: e.UpdatedAt,
    dataRecordsEnabled: e.DataRecordsEnabled,
    entityType: e.EntityType,
    tags: (e.Tags ?? []).map((t) => ({ id: t.ID, value: t.Value })),
    variants: (e.Variants ?? []).map(mapVariant),
    segments: (e.Segments ?? []).map(mapSegment),
  }
}

/**
 * Mapped dump shared across pages so list → detail → back is one fetch,
 * not three. `fetchFlags` refreshes it; `getFlags` reads through it.
 */
let dumpCache: Flag[] | null = null
let inFlight: Promise<ApiResult<Flag[]>> | null = null

export function clearDumpCache(): void {
  dumpCache = null
}

/**
 * Always refetch the export and refresh the cache (used on list mount).
 * Concurrent callers (e.g. getFlag + listAllTags on a detail deep link)
 * share one request.
 */
export function fetchFlags(): Promise<ApiResult<Flag[]>> {
  if (inFlight) return inFlight
  inFlight = requestJson<EvalCacheDump>({
    method: 'GET',
    path: '/export/eval_cache/json',
  })
    .then((res) => {
      if (!res.ok) return res
      const flags = (res.value?.Flags ?? []).map(mapExportedFlag)
      // The export iterates a map — order is random per fetch. Sort by id for a
      // stable list; callers reverse for newest-first display like the DB mode.
      flags.sort((a, b) => (a.id ?? 0) - (b.id ?? 0))
      dumpCache = flags
      return ok(flags)
    })
    .finally(() => {
      inFlight = null
    })
  return inFlight
}

/** Read through the cache; fetch only when it is empty (e.g. detail deep link). */
export async function getFlags(): Promise<ApiResult<Flag[]>> {
  if (dumpCache) return ok(dumpCache)
  return fetchFlags()
}

export async function getFlag(flagId: string | number): Promise<ApiResult<Flag>> {
  const res = await getFlags()
  if (!res.ok) return res
  const id = Number(flagId)
  const found = res.value.find((f) => f.id === id)
  if (!found) {
    return err(new ApiHttpError(404, `unable to find flag ${flagId} in the eval cache`))
  }
  return ok(found)
}

/**
 * Unique tags across the dump. In the JSON source the same tag value on two
 * flags is two entities — dedupe by value to mirror the DB's many2many table.
 */
export async function listAllTags(): Promise<ApiResult<Tag[]>> {
  const res = await getFlags()
  if (!res.ok) return res
  const seen = new Set<string>()
  const tags: Tag[] = []
  for (const f of res.value) {
    for (const t of f.tags ?? []) {
      if (seen.has(t.value)) continue
      seen.add(t.value)
      tags.push(t)
    }
  }
  return ok(tags)
}
