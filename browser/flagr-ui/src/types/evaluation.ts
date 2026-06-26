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