import type { Health } from './types'
import type { ApiResult } from './result'
import { requestJson } from './http'

export const getHealth = (): Promise<ApiResult<Health>> =>
  requestJson<Health>({ method: 'GET', path: '/health' })
