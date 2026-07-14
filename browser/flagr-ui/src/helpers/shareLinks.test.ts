import { describe, expect, it } from 'vitest'
import {
  FLAG_TAB_CONFIG,
  FLAG_TAB_HISTORY,
  buildAppUrl,
  flagHistoryUrl,
  flagSnapshotUrl,
  flagUrl,
  parseFlagDeepLink,
  snapshotElementId,
} from './shareLinks'

const loc = {
  origin: 'https://flagr.example.com',
  pathname: '/ui/',
  search: '',
}

describe('shareLinks', () => {
  it('buildAppUrl joins origin, pathname, search, and hash path', () => {
    expect(buildAppUrl('/flags/42', loc)).toBe('https://flagr.example.com/ui/#/flags/42')
    expect(
      buildAppUrl('/flags/42', { ...loc, search: '?embed=1' }),
    ).toBe('https://flagr.example.com/ui/?embed=1#/flags/42')
  })

  it('flagUrl is absolute Config deep link', () => {
    expect(flagUrl(42, loc)).toBe('https://flagr.example.com/ui/#/flags/42')
    expect(flagUrl('7', loc)).toBe('https://flagr.example.com/ui/#/flags/7')
  })

  it('flagHistoryUrl includes tab=history', () => {
    expect(flagHistoryUrl(42, loc)).toBe(
      'https://flagr.example.com/ui/#/flags/42?tab=history',
    )
  })

  it('flagSnapshotUrl includes tab and snapshot', () => {
    expect(flagSnapshotUrl(42, 87, loc)).toBe(
      'https://flagr.example.com/ui/#/flags/42?tab=history&snapshot=87',
    )
  })

  it('snapshotElementId is stable', () => {
    expect(snapshotElementId(87)).toBe('snapshot-87')
  })

  it('parseFlagDeepLink defaults to config', () => {
    expect(parseFlagDeepLink(undefined)).toEqual({
      tab: FLAG_TAB_CONFIG,
      snapshotId: null,
    })
    expect(parseFlagDeepLink({})).toEqual({
      tab: FLAG_TAB_CONFIG,
      snapshotId: null,
    })
  })

  it('parseFlagDeepLink reads history tab and snapshot', () => {
    expect(parseFlagDeepLink({ tab: 'history' })).toEqual({
      tab: FLAG_TAB_HISTORY,
      snapshotId: null,
    })
    expect(parseFlagDeepLink({ tab: 'history', snapshot: '87' })).toEqual({
      tab: FLAG_TAB_HISTORY,
      snapshotId: 87,
    })
    // Snapshot alone implies history
    expect(parseFlagDeepLink({ snapshot: '3' })).toEqual({
      tab: FLAG_TAB_HISTORY,
      snapshotId: 3,
    })
  })

  it('parseFlagDeepLink ignores invalid snapshot values', () => {
    expect(parseFlagDeepLink({ tab: 'history', snapshot: 'nope' })).toEqual({
      tab: FLAG_TAB_HISTORY,
      snapshotId: null,
    })
    expect(parseFlagDeepLink({ snapshot: '0' })).toEqual({
      tab: FLAG_TAB_CONFIG,
      snapshotId: null,
    })
    expect(parseFlagDeepLink({ snapshot: ['9', '10'] })).toEqual({
      tab: FLAG_TAB_HISTORY,
      snapshotId: 9,
    })
  })
})
