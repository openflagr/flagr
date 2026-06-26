import { Effect } from 'effect'
import type { BatchEvalContext, BatchEvalResult, EvalContext, EvalResult } from '@/types'
import type { ApiError } from './errors'
import { requestJson } from './http'

export const postEvaluation = (
  body: EvalContext,
): Effect.Effect<EvalResult, ApiError> =>
  requestJson<EvalResult>({ method: 'POST', path: '/evaluation', body })

export const postEvaluationBatch = (
  body: BatchEvalContext,
): Effect.Effect<BatchEvalResult, ApiError> =>
  requestJson<BatchEvalResult>({ method: 'POST', path: '/evaluation/batch', body })