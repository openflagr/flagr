import type { BatchEvalContext, BatchEvalResult, EvalContext, EvalResult } from '@/api/types'

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