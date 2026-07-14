/** Deep-link query keys for the flag detail page (hash router). */
export const FLAG_QUERY_TAB = 'tab'
export const FLAG_QUERY_SNAPSHOT = 'snapshot'

export const FLAG_TAB_CONFIG = 'config'
export const FLAG_TAB_HISTORY = 'history'

export type FlagTabName = typeof FLAG_TAB_CONFIG | typeof FLAG_TAB_HISTORY

/** DOM id for a history snapshot card (scroll target). */
export function snapshotElementId(snapshotId: string | number): string {
  return `snapshot-${snapshotId}`
}

/** Minimal location surface so helpers stay unit-testable without a real window. */
export type LocationLike = Pick<Location, 'origin' | 'pathname' | 'search'>

export function buildAppUrl(hashPath: string, loc: LocationLike): string {
  const path = hashPath.startsWith('/') ? hashPath : `/${hashPath}`
  return `${loc.origin}${loc.pathname}${loc.search}#${path}`
}

/** Absolute UI URL for a flag (Config tab — default). */
export function flagUrl(flagId: string | number, loc: LocationLike): string {
  return buildAppUrl(`/flags/${flagId}`, loc)
}

/** Absolute UI URL for a flag History tab. */
export function flagHistoryUrl(flagId: string | number, loc: LocationLike): string {
  return buildAppUrl(`/flags/${flagId}?${FLAG_QUERY_TAB}=${FLAG_TAB_HISTORY}`, loc)
}

/** Absolute UI URL for a specific history snapshot block. */
export function flagSnapshotUrl(
  flagId: string | number,
  snapshotId: string | number,
  loc: LocationLike,
): string {
  const q = new URLSearchParams({
    [FLAG_QUERY_TAB]: FLAG_TAB_HISTORY,
    [FLAG_QUERY_SNAPSHOT]: String(snapshotId),
  })
  return buildAppUrl(`/flags/${flagId}?${q.toString()}`, loc)
}

export interface FlagDeepLink {
  tab: FlagTabName
  snapshotId: number | null
}

/** Parse flag deep-link intent from vue-router query (values may be string | string[]). */
export function parseFlagDeepLink(
  query: Record<string, unknown> | undefined | null,
): FlagDeepLink {
  if (!query) {
    return { tab: FLAG_TAB_CONFIG, snapshotId: null }
  }

  const rawTab = firstQueryValue(query[FLAG_QUERY_TAB])
  const rawSnapshot = firstQueryValue(query[FLAG_QUERY_SNAPSHOT])

  let snapshotId: number | null = null
  if (rawSnapshot != null && rawSnapshot !== '') {
    const n = Number(rawSnapshot)
    if (Number.isFinite(n) && n > 0) {
      snapshotId = n
    }
  }

  if (rawTab === FLAG_TAB_HISTORY || snapshotId != null) {
    return { tab: FLAG_TAB_HISTORY, snapshotId }
  }

  return { tab: FLAG_TAB_CONFIG, snapshotId: null }
}

function firstQueryValue(value: unknown): string | null {
  if (value == null) return null
  if (Array.isArray(value)) {
    const first = value[0]
    return first == null ? null : String(first)
  }
  return String(value)
}
