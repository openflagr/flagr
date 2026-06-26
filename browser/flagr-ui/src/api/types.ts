/** Mirrors swagger_gen/models — UI uses relaxed optionality for JSON payloads. */

export interface Tag {
  id?: number
  value: string
}

export interface IdentifiedTag extends Tag {
  id: number
}

export interface Distribution {
  id?: number
  percent: number
  variantID: number
  variantKey: string
}

/** Distribution row while editing (server id omitted on save). */
export type DistributionDraft = Omit<Distribution, 'id'> & { bitmap?: string }

export interface Constraint {
  id?: number
  operator: string
  property: string
  value: string
}

export interface IdentifiedConstraint extends Constraint {
  id: number
}

export interface Segment {
  id?: number
  description: string
  rolloutPercent: number
  rank?: number
  constraints?: Constraint[]
  distributions?: Distribution[]
}

export interface IdentifiedSegment extends Segment {
  id: number
}

export interface Variant {
  id?: number
  key: string
  attachment?: Record<string, unknown> | string
  attachmentValid?: boolean
}

export interface IdentifiedVariant extends Variant {
  id: number
}

export interface Flag {
  id?: number
  description: string
  key?: string
  enabled?: boolean
  dataRecordsEnabled?: boolean
  entityType?: string
  notes?: string
  createdBy?: string
  updatedBy?: string
  updatedAt?: string
  tags?: Tag[]
  variants: Variant[]
  segments?: Segment[]
}

/** Flag after `normalizeFlag` (empty arrays materialized; variants may carry UI validation state). */
export type FlagView = Omit<Flag, 'variants'> & { variants: Variant[] }

export interface CreateFlagPayload {
  description: string
  template?: string
  [key: string]: unknown
}

export interface UpdateFlagPayload {
  description: string
  dataRecordsEnabled?: boolean
  key: string
  entityType: string
  notes: string
}

export interface FlagSnapshot {
  id: number
  updatedAt?: string
  updatedBy?: string
  flag: Flag
}

export interface SnapshotMaxId {
  maxID: number
}

export function requireSegmentId(segment: Segment): number {
  if (segment.id == null) throw new Error('segment id required')
  return segment.id
}

export function requireVariantId(variant: Variant): number {
  if (variant.id == null) throw new Error('variant id required')
  return variant.id
}

export function requireConstraintId(constraint: Constraint): number {
  if (constraint.id == null) throw new Error('constraint id required')
  return constraint.id
}

export function requireTagId(tag: Tag): number {
  if (tag.id == null) throw new Error('tag id required')
  return tag.id
}

export function requireFlagId(flag: Flag): number {
  if (flag.id == null) throw new Error('flag id required')
  return flag.id
}

export function pluckSegmentIds(segments: Segment[]): number[] {
  return segments.map((s) => requireSegmentId(s))
}
export interface EvalContext {
  entityID?: string
  entityType?: string
  entityContext?: Record<string, unknown>
  enableDebug?: boolean
  flagID?: number
  flagKey?: string
}

export interface BatchEvalContext {
  entities?: EvalContext[]
  enableDebug?: boolean
  flagIDs?: number[]
}

export type EvalResult = Record<string, unknown>
export type BatchEvalResult = Record<string, unknown>

export interface EvalSummaryConstraint {
  constraintID: unknown
  constraintProperty: unknown
  constraintOperator: unknown
  constraintValue: unknown
  matched: unknown
}

export interface EvalSummarySegment {
  segmentID: unknown
  description: unknown
  rolloutPercent: unknown
  matched: unknown
  msg: unknown
  constraints: EvalSummaryConstraint[]
}

export interface EvalSummary {
  variantKey: string
  variantID: unknown
  segments: EvalSummarySegment[]
}
export type OperatorValue =
  | 'EQ' | 'NEQ' | 'LT' | 'LTE' | 'GT' | 'GTE'
  | 'EREG' | 'NEREG' | 'IN' | 'NOTIN' | 'CONTAINS' | 'NOTCONTAINS'
