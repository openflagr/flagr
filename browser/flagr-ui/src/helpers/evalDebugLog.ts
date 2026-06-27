import type { EvalResult, EvalSummary } from '@/api/types'

/** Parse evaluation debug payload from POST /evaluation (untrusted API shape). */
export function evalSummaryFromResult(result: EvalResult): EvalSummary | null {
  if (!result?.evalDebugLog) return null
  const log = result.evalDebugLog as Record<string, unknown>
  const segments = ((log.segmentDebugLogs as unknown[]) || []).map((s) => {
    const seg = s as Record<string, unknown>
    return {
      segmentID: seg.segmentID,
      description: seg.description,
      rolloutPercent: seg.rolloutPercent,
      matched: seg.matched,
      msg: seg.msg,
      constraints: ((seg.constraintDebugLogs as unknown[]) || []).map((c) => {
        const con = c as Record<string, unknown>
        return {
          constraintID: con.constraintID,
          constraintProperty: con.constraintProperty,
          constraintOperator: con.constraintOperator,
          constraintValue: con.constraintValue,
          matched: con.matched,
        }
      }),
    }
  })
  return {
    variantKey: (result.variantKey as string) || '—',
    variantID: result.variantID,
    segments,
  }
}