import type { EvalDebugLog, EvalResult, EvalSummary, EvalSummarySegment } from '@/api/types'

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