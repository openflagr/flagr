import type { BatchEvalContext, BatchEvalResult, EvalContext, EvalResult } from './types'
import type { ApiResult } from './result'
import { requestJson } from './http'

export const postEvaluation = (body: EvalContext): Promise<ApiResult<EvalResult>> =>
  requestJson<EvalResult>({ method: 'POST', path: '/evaluation', body })

export const postEvaluationBatch = (
  body: BatchEvalContext,
): Promise<ApiResult<BatchEvalResult>> =>
  requestJson<BatchEvalResult>({ method: 'POST', path: '/evaluation/batch', body })