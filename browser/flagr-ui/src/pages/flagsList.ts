import type { Flag } from '@/api/types'

export interface FlagsCache {
  flags: Flag[]
  maxSnapshotID: number
}

let flagsCache: FlagsCache | null = null

export function getFlagsCache(): FlagsCache | null {
  return flagsCache
}

export function setFlagsCache(cache: FlagsCache): void {
  flagsCache = cache
}