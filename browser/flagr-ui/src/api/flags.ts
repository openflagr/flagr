import { Effect } from 'effect'
import type {
  Constraint,
  CreateFlagPayload,
  Distribution,
  Flag,
  FlagSnapshot,
  Segment,
  SnapshotMaxId,
  Tag,
  UpdateFlagPayload,
  Variant,
} from './types'
import type { ApiError } from './errors'
import { requestJson, requestVoid } from './http'

type FlagId = string | number

const flag = (id: FlagId) => `/flags/${id}`

const get = <T>(path: string) => requestJson<T>({ method: 'GET', path })
const post = <T>(path: string, body: unknown) => requestJson<T>({ method: 'POST', path, body })
const putVoid = (path: string, body?: unknown) => requestVoid({ method: 'PUT', path, body })
const del = (path: string) => requestVoid({ method: 'DELETE', path })

export const listFlags = (): Effect.Effect<Flag[], ApiError> => get('/flags')

export const getSnapshotMaxId = (): Effect.Effect<SnapshotMaxId, ApiError> =>
  get('/flags/snapshots/max_id')

export const listFlagsIfStale = Effect.fn('flags.listFlagsIfStale')(function* (
  cachedMaxId: number | undefined,
) {
  const { maxID } = yield* getSnapshotMaxId()
  if (cachedMaxId !== undefined && maxID === cachedMaxId) {
    return null
  }
  const flags = yield* listFlags()
  return { flags: [...flags].reverse(), maxSnapshotID: maxID }
})

export const listDeletedFlags = (): Effect.Effect<Flag[], ApiError> =>
  get('/flags?deleted=true')

export const createFlag = (body: CreateFlagPayload): Effect.Effect<Flag, ApiError> =>
  post('/flags', body)

export const restoreFlag = (flagId: number): Effect.Effect<Flag, ApiError> =>
  requestJson<Flag>({ method: 'PUT', path: `${flag(flagId)}/restore` })

export const getFlag = (flagId: FlagId): Effect.Effect<Flag, ApiError> => get(flag(flagId))

export const updateFlag = (flagId: FlagId, body: UpdateFlagPayload): Effect.Effect<void, ApiError> =>
  putVoid(flag(flagId), body)

export const setFlagEnabled = (flagId: FlagId, enabled: boolean): Effect.Effect<void, ApiError> =>
  putVoid(`${flag(flagId)}/enabled`, { enabled })

export const deleteFlag = (flagId: FlagId): Effect.Effect<void, ApiError> => del(flag(flagId))

export const listAllTags = (): Effect.Effect<Tag[], ApiError> => get('/tags')

export const createTag = (flagId: FlagId, value: string): Effect.Effect<Tag, ApiError> =>
  post(`${flag(flagId)}/tags`, { value })

export const deleteTag = (flagId: FlagId, tagId: number): Effect.Effect<void, ApiError> =>
  del(`${flag(flagId)}/tags/${tagId}`)

export const createVariant = (flagId: FlagId, key: string): Effect.Effect<Variant, ApiError> =>
  post(`${flag(flagId)}/variants`, { key })

export const updateVariant = (
  flagId: FlagId,
  variantId: number,
  body: { key: string; attachment?: unknown },
): Effect.Effect<void, ApiError> => putVoid(`${flag(flagId)}/variants/${variantId}`, body)

export const deleteVariant = (flagId: FlagId, variantId: number): Effect.Effect<void, ApiError> =>
  del(`${flag(flagId)}/variants/${variantId}`)

export const createSegment = (
  flagId: FlagId,
  body: { description: string; rolloutPercent: number },
): Effect.Effect<Segment, ApiError> => post(`${flag(flagId)}/segments`, body)

export const updateSegment = (
  flagId: FlagId,
  segmentId: number,
  body: { description: string; rolloutPercent: number },
): Effect.Effect<void, ApiError> => putVoid(`${flag(flagId)}/segments/${segmentId}`, body)

export const deleteSegment = (flagId: FlagId, segmentId: number): Effect.Effect<void, ApiError> =>
  del(`${flag(flagId)}/segments/${segmentId}`)

export const reorderSegments = (flagId: FlagId, segmentIDs: number[]): Effect.Effect<void, ApiError> =>
  putVoid(`${flag(flagId)}/segments/reorder`, { segmentIDs })

export const createConstraint = (
  flagId: FlagId,
  segmentId: number,
  body: Constraint,
): Effect.Effect<Constraint, ApiError> =>
  post(`${flag(flagId)}/segments/${segmentId}/constraints`, body)

export const updateConstraint = (
  flagId: FlagId,
  segmentId: number,
  constraintId: number,
  body: Constraint,
): Effect.Effect<void, ApiError> =>
  putVoid(`${flag(flagId)}/segments/${segmentId}/constraints/${constraintId}`, body)

export const deleteConstraint = (
  flagId: FlagId,
  segmentId: number,
  constraintId: number,
): Effect.Effect<void, ApiError> =>
  del(`${flag(flagId)}/segments/${segmentId}/constraints/${constraintId}`)

export const putSegmentDistributions = (
  flagId: FlagId,
  segmentId: number,
  distributions: Distribution[],
): Effect.Effect<Distribution[], ApiError> =>
  requestJson<Distribution[]>({
    method: 'PUT',
    path: `${flag(flagId)}/segments/${segmentId}/distributions`,
    body: { distributions },
  })

export const listFlagSnapshots = (flagId: FlagId): Effect.Effect<FlagSnapshot[], ApiError> =>
  get(`${flag(flagId)}/snapshots`)

export const listEntityTypes = (): Effect.Effect<string[], ApiError> =>
  get('/flags/entity_types')

export interface FlagPageLoad {
  flag: Flag
  allTags: Tag[]
  /** From `GET /flags/entity_types` (page may override with env). */
  entityTypesFromApi: string[]
}

export const loadFlagPageContext = Effect.fn('flags.loadFlagPageContext')(function* (
  flagId: FlagId,
) {
  const [flag, allTags, entityTypesFromApi] = yield* Effect.all(
    [getFlag(flagId), listAllTags(), listEntityTypes()],
    { concurrency: 'unbounded' },
  )
  return { flag, allTags, entityTypesFromApi } satisfies FlagPageLoad
})

export const loadFlagAndAllTags = Effect.fn('flags.loadFlagAndAllTags')(function* (flagId: FlagId) {
  const [flag, allTags] = yield* Effect.all(
    [getFlag(flagId), listAllTags()],
    { concurrency: 'unbounded' },
  )
  return { flag, allTags }
})