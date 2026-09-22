import type {
  BatchEvalContext,
  BatchEvalResult,
  EvalContext,
  EvalDebugLog,
  EvalResult,
  EvalSummary,
  EvalSummarySegment,
} from '@/api/types'

export function parseEvalContextJson(text: string): EvalContext | null {
  try {
    return JSON.parse(text) as EvalContext
  } catch {
    return null
  }
}

export function parseEvalResultJson(text: string): EvalResult | null {
  try {
    return JSON.parse(text) as EvalResult
  } catch {
    return null
  }
}

export function parseBatchEvalContextJson(text: string): BatchEvalContext | null {
  try {
    return JSON.parse(text) as BatchEvalContext
  } catch {
    return null
  }
}

export function parseBatchEvalResultJson(text: string): BatchEvalResult | null {
  try {
    return JSON.parse(text) as BatchEvalResult
  } catch {
    return null
  }
}

/** Parse evaluation debug payload from POST /evaluation (swagger evalDebugLog). */
export function evalSummaryFromResult(result: EvalResult): EvalSummary | null {
  if (result.flagID == null && result.flagKey == null) return null

  const log: EvalDebugLog | undefined = result.evalDebugLog
  const segments: EvalSummarySegment[] = (log?.segmentDebugLogs ?? []).map((seg) => ({
    segmentID: seg.segmentID,
    msg: seg.msg,
    constraints: [],
  }))

  return {
    variantKey: result.variantKey ?? '—',
    variantID: result.variantID,
    segments,
  }
}
