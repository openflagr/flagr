import constants from '@/helpers/constants'
import { setEntityTypeOptionsFromApiKeys, type EntityTypeOption } from '@/helpers/flagModel'
import type {
  Constraint,
  CreateFlagPayload,
  DuplicateFlagPayload,
  Distribution,
  Flag,
  FlagSnapshot,
  PutVariantBody,
  Segment,
  SnapshotMaxId,
  Tag,
  UpdateFlagPayload,
  Variant,
} from './types'
import type { ApiResult } from './result'
import { ok } from './result'
import { requestJson, requestVoid } from './http'
import * as evalCache from './evalCache'

type FlagId = string | number

const flag = (id: FlagId) => `/flags/${id}`

const get = <T>(path: string) => requestJson<T>({ method: 'GET', path })
const post = <T>(path: string, body: unknown) => requestJson<T>({ method: 'POST', path, body })
const putVoid = (path: string, body?: unknown) => requestVoid({ method: 'PUT', path, body })
const del = (path: string) => requestVoid({ method: 'DELETE', path })

export const listFlags = (): Promise<ApiResult<Flag[]>> => get('/flags')

export const getSnapshotMaxId = (): Promise<ApiResult<SnapshotMaxId>> =>
  get('/flags/snapshots/max_id')

export type FlagReadSource = 'http' | 'evalCache'

interface FlagReads {
  listFlagsIfStale: (
    cachedMaxId: number | undefined,
  ) => Promise<ApiResult<{ flags: Flag[]; maxSnapshotID: number } | null>>
  getFlag: (flagId: FlagId) => Promise<ApiResult<Flag>>
  listAllTags: () => Promise<ApiResult<Tag[]>>
  listDeletedFlags: () => Promise<ApiResult<Flag[]>>
  listFlagSnapshots: (flagId: FlagId) => Promise<ApiResult<FlagSnapshot[]>>
  listEntityTypes: () => Promise<ApiResult<string[]>>
}

async function listFlagsIfStaleFromHTTP(
  cachedMaxId: number | undefined,
): Promise<ApiResult<{ flags: Flag[]; maxSnapshotID: number } | null>> {
  const maxRes = await getSnapshotMaxId()
  if (!maxRes.ok) return maxRes
  const { maxID } = maxRes.value
  if (cachedMaxId !== undefined && maxID === cachedMaxId) {
    return ok(null)
  }
  const flagsRes = await listFlags()
  if (!flagsRes.ok) return flagsRes
  return ok({ flags: [...flagsRes.value].reverse(), maxSnapshotID: maxID })
}

async function listFlagsIfStaleFromEvalCache(): Promise<
  ApiResult<{ flags: Flag[]; maxSnapshotID: number } | null>
> {
  // No snapshots in eval-only mode — refetch the export dump on every list
  // mount so JSON source changes show up without a change token.
  const flagsRes = await evalCache.fetchFlags()
  if (!flagsRes.ok) return flagsRes
  return ok({ flags: [...flagsRes.value].reverse(), maxSnapshotID: 0 })
}

const httpReads: FlagReads = {
  listFlagsIfStale: listFlagsIfStaleFromHTTP,
  getFlag: (flagId) => get(flag(flagId)),
  listAllTags: () => get('/tags'),
  listDeletedFlags: () => get('/flags?deleted=true'),
  listFlagSnapshots: (flagId) => get(`${flag(flagId)}/snapshots`),
  listEntityTypes: () => get('/flags/entity_types'),
}

const evalCacheReads: FlagReads = {
  listFlagsIfStale: listFlagsIfStaleFromEvalCache,
  getFlag: evalCache.getFlag,
  listAllTags: evalCache.listAllTags,
  // JSON sources have no soft-deletes; history lives in Git; entity types
  // are recorded into the DB on evaluation — none of that exists here.
  listDeletedFlags: () => Promise.resolve(ok([])),
  listFlagSnapshots: () => Promise.resolve(ok([])),
  listEntityTypes: () => Promise.resolve(ok([])),
}

let reads: FlagReads = httpReads

/** Swap the flag read plane. Called once when the server mode is known. */
export function setFlagReadSource(source: FlagReadSource): void {
  reads = source === 'evalCache' ? evalCacheReads : httpReads
}

export function getFlagReadSource(): FlagReadSource {
  return reads === evalCacheReads ? 'evalCache' : 'http'
}

export const listFlagsIfStale = (
  cachedMaxId: number | undefined,
): Promise<ApiResult<{ flags: Flag[]; maxSnapshotID: number } | null>> =>
  reads.listFlagsIfStale(cachedMaxId)

export const listDeletedFlags = (): Promise<ApiResult<Flag[]>> => reads.listDeletedFlags()

export const createFlag = (body: CreateFlagPayload): Promise<ApiResult<Flag>> => post('/flags', body)

export const restoreFlag = (flagId: number): Promise<ApiResult<Flag>> =>
  requestJson<Flag>({ method: 'PUT', path: `${flag(flagId)}/restore` })

export const getFlag = (flagId: FlagId): Promise<ApiResult<Flag>> => reads.getFlag(flagId)

export const duplicateFlag = (
  flagId: FlagId,
  body: DuplicateFlagPayload = {},
): Promise<ApiResult<Flag>> =>
  post(`${flag(flagId)}/duplicate`, body)

export const updateFlag = (
  flagId: FlagId,
  body: UpdateFlagPayload,
): Promise<ApiResult<void>> => putVoid(flag(flagId), body)

export const setFlagEnabled = (
  flagId: FlagId,
  enabled: boolean,
): Promise<ApiResult<void>> => putVoid(`${flag(flagId)}/enabled`, { enabled })

export const deleteFlag = (flagId: FlagId): Promise<ApiResult<void>> => del(flag(flagId))

export const listAllTags = (): Promise<ApiResult<Tag[]>> => reads.listAllTags()

export const createTag = (flagId: FlagId, value: string): Promise<ApiResult<Tag>> =>
  post(`${flag(flagId)}/tags`, { value })

export async function createTagAndRefreshAllTags(
  flagId: FlagId,
  value: string,
): Promise<ApiResult<{ tag: Tag; allTags: Tag[] }>> {
  const tagRes = await createTag(flagId, value)
  if (!tagRes.ok) return tagRes
  const allTagsRes = await listAllTags()
  if (!allTagsRes.ok) return allTagsRes
  return ok({ tag: tagRes.value, allTags: allTagsRes.value })
}

export const deleteTag = (flagId: FlagId, tagId: number): Promise<ApiResult<void>> =>
  del(`${flag(flagId)}/tags/${tagId}`)

export const createVariant = (flagId: FlagId, key: string): Promise<ApiResult<Variant>> =>
  post(`${flag(flagId)}/variants`, { key })

export const updateVariant = (
  flagId: FlagId,
  variantId: number,
  body: PutVariantBody,
): Promise<ApiResult<void>> => putVoid(`${flag(flagId)}/variants/${variantId}`, body)

export const deleteVariant = (flagId: FlagId, variantId: number): Promise<ApiResult<void>> =>
  del(`${flag(flagId)}/variants/${variantId}`)

export const createSegment = (
  flagId: FlagId,
  body: { description: string; rolloutPercent: number },
): Promise<ApiResult<Segment>> => post(`${flag(flagId)}/segments`, body)

export const updateSegment = (
  flagId: FlagId,
  segmentId: number,
  body: { description: string; rolloutPercent: number },
): Promise<ApiResult<void>> => putVoid(`${flag(flagId)}/segments/${segmentId}`, body)

export const deleteSegment = (flagId: FlagId, segmentId: number): Promise<ApiResult<void>> =>
  del(`${flag(flagId)}/segments/${segmentId}`)

export const reorderSegments = (
  flagId: FlagId,
  segmentIDs: number[],
): Promise<ApiResult<void>> => putVoid(`${flag(flagId)}/segments/reorder`, { segmentIDs })

export const createConstraint = (
  flagId: FlagId,
  segmentId: number,
  body: Constraint,
): Promise<ApiResult<Constraint>> =>
  post(`${flag(flagId)}/segments/${segmentId}/constraints`, body)

export const updateConstraint = (
  flagId: FlagId,
  segmentId: number,
  constraintId: number,
  body: Constraint,
): Promise<ApiResult<void>> =>
  putVoid(`${flag(flagId)}/segments/${segmentId}/constraints/${constraintId}`, body)

export const deleteConstraint = (
  flagId: FlagId,
  segmentId: number,
  constraintId: number,
): Promise<ApiResult<void>> =>
  del(`${flag(flagId)}/segments/${segmentId}/constraints/${constraintId}`)

export const putSegmentDistributions = (
  flagId: FlagId,
  segmentId: number,
  distributions: Distribution[],
): Promise<ApiResult<Distribution[]>> =>
  requestJson<Distribution[]>({
    method: 'PUT',
    path: `${flag(flagId)}/segments/${segmentId}/distributions`,
    body: { distributions },
  })

export const listFlagSnapshots = (flagId: FlagId): Promise<ApiResult<FlagSnapshot[]>> =>
  reads.listFlagSnapshots(flagId)

export const listEntityTypes = (): Promise<ApiResult<string[]>> => reads.listEntityTypes()

export interface FlagPageLoad {
  flag: Flag
  allTags: Tag[]
  entityTypes: EntityTypeOption[]
  allowCreateEntityType: boolean
}

const skipEntityTypesApi = !!constants.FLAGR_UI_POSSIBLE_ENTITY_TYPES

export async function loadFlagPageContext(flagId: FlagId): Promise<ApiResult<FlagPageLoad>> {
  if (skipEntityTypesApi) {
    const [flagRes, allTagsRes] = await Promise.all([getFlag(flagId), listAllTags()])
    if (!flagRes.ok) return flagRes
    if (!allTagsRes.ok) return allTagsRes
    const { entityTypes, allowCreateEntityType } = setEntityTypeOptionsFromApiKeys([])
    return ok({
      flag: flagRes.value,
      allTags: allTagsRes.value,
      entityTypes,
      allowCreateEntityType,
    })
  }
  const [flagRes, allTagsRes, entityTypesRes] = await Promise.all([
    getFlag(flagId),
    listAllTags(),
    listEntityTypes(),
  ])
  if (!flagRes.ok) return flagRes
  if (!allTagsRes.ok) return allTagsRes
  if (!entityTypesRes.ok) return entityTypesRes
  const { entityTypes, allowCreateEntityType } = setEntityTypeOptionsFromApiKeys(
    entityTypesRes.value,
  )
  return ok({
    flag: flagRes.value,
    allTags: allTagsRes.value,
    entityTypes,
    allowCreateEntityType,
  })
}

export async function loadFlagAndAllTags(
  flagId: FlagId,
): Promise<ApiResult<{ flag: Flag; allTags: Tag[] }>> {
  const [flagRes, allTagsRes] = await Promise.all([getFlag(flagId), listAllTags()])
  if (!flagRes.ok) return flagRes
  if (!allTagsRes.ok) return allTagsRes
  return ok({ flag: flagRes.value, allTags: allTagsRes.value })
}

export async function deleteTagAndReload(
  flagId: FlagId,
  tagId: number,
): Promise<ApiResult<{ flag: Flag; allTags: Tag[] }>> {
  const delRes = await deleteTag(flagId, tagId)
  if (!delRes.ok) return delRes
  return loadFlagAndAllTags(flagId)
}