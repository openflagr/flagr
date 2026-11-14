import { describe, expect, it } from 'vitest'
import type { Enricher } from '@/api/types'
import { enricherPropertyGroups, enrichedProperties } from './enricherOptions'

const catalog: Enricher[] = [
  { namespace: 'ts', scope: 'global', properties: ['@ts_hour', '@ts'] },
  { namespace: 'http', scope: 'global', properties: ['@http_x_env'] },
  { namespace: 'jev', scope: 'flag', properties: ['@jev_plan_tier', '@jev_churn'] },
]

describe('enricherPropertyGroups', () => {
  it('groups by namespace and sorts options', () => {
    expect(enricherPropertyGroups(catalog)).toEqual([
      { namespace: 'ts', label: 'ts', options: ['@ts', '@ts_hour'] },
      { namespace: 'http', label: 'http', options: ['@http_x_env'] },
      { namespace: 'jev', label: 'jev', options: ['@jev_churn', '@jev_plan_tier'] },
    ])
  })

  it('drops enrichers without properties', () => {
    expect(enricherPropertyGroups([{ namespace: 'http', scope: 'global', properties: [] }])).toEqual([])
  })

  it('handles undefined and empty catalogs', () => {
    expect(enricherPropertyGroups(undefined)).toEqual([])
    expect(enrichedProperties(undefined)).toEqual([])
  })
})

describe('enrichedProperties', () => {
  it('flattens the catalog in order', () => {
    expect(enrichedProperties(catalog)).toEqual([
      '@ts',
      '@ts_hour',
      '@http_x_env',
      '@jev_churn',
      '@jev_plan_tier',
    ])
  })
})
