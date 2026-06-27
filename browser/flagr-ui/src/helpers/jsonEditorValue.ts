import type { BatchEvalResult, EvalResult } from '@/api/types'

/** vue3-ts-jsoneditor emits untyped JSON; narrow at the component edge. */
export function asJsonObject(value: unknown): Record<string, unknown> | null {
  if (value !== null && typeof value === 'object' && !Array.isArray(value)) {
    return value as Record<string, unknown>
  }
  return null
}

export function asBatchEvalResult(value: unknown): BatchEvalResult | null {
  const o = asJsonObject(value)
  if (!o || !Array.isArray(o.evaluationResults)) return null
  return { evaluationResults: o.evaluationResults as EvalResult[] }
}