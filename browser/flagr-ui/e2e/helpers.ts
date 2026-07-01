import type { Constraint, Distribution, Flag, FlagSnapshot, Segment, Tag, Variant } from '../src/api/types'

export const API = process.env.API_URL || 'http://localhost:18000/api/v1'

export interface CreateFlagOpts {
  description?: string
}

export async function createFlag(opts: CreateFlagOpts = {}): Promise<Flag> {
  const r = await fetch(`${API}/flags`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ description: opts.description || `e2e-${Date.now()}` }),
  })
  if (!r.ok) throw new Error(`createFlag failed: ${r.status} ${await r.text()}`)
  return r.json() as Promise<Flag>
}

export async function createFlagWithVariants(opts: CreateFlagOpts = {}): Promise<Flag> {
  const flag = await createFlag(opts)
  for (const key of ['control', 'test']) {
    const r = await fetch(`${API}/flags/${flag.id}/variants`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ key }),
    })
    if (!r.ok) throw new Error(`createVariant(${key}) failed: ${r.status}`)
  }
  return flag
}

export async function deleteFlag(flagId: number): Promise<void> {
  const r = await fetch(`${API}/flags/${flagId}`, { method: 'DELETE' })
  if (!r.ok) throw new Error(`deleteFlag(${flagId}) failed: ${r.status}`)
}

export async function createSegment(flagId: number, desc: string): Promise<Segment> {
  const r = await fetch(`${API}/flags/${flagId}/segments`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ description: desc, rolloutPercent: 100 }),
  })
  if (!r.ok) throw new Error(`createSegment failed: ${r.status}`)
  return r.json() as Promise<Segment>
}

export async function getFlag(flagId: number): Promise<Flag> {
  const r = await fetch(`${API}/flags/${flagId}`)
  if (!r.ok) throw new Error(`getFlag(${flagId}) failed: ${r.status}`)
  return r.json() as Promise<Flag>
}

export async function createTag(flagId: number, value: string): Promise<Tag> {
  const r = await fetch(`${API}/flags/${flagId}/tags`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ value }),
  })
  if (!r.ok) throw new Error(`createTag failed: ${r.status}`)
  return r.json() as Promise<Tag>
}

export async function putSegmentDistributions(
  flagId: number,
  segmentId: number,
  distributions: Distribution[],
): Promise<Distribution[]> {
  const r = await fetch(`${API}/flags/${flagId}/segments/${segmentId}/distributions`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ distributions }),
  })
  if (!r.ok) throw new Error(`putSegmentDistributions failed: ${r.status} ${await r.text()}`)
  return r.json() as Promise<Distribution[]>
}

export async function createConstraint(flagId: number, segId: number): Promise<Constraint> {
  const r = await fetch(`${API}/flags/${flagId}/segments/${segId}/constraints`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ property: 'country', operator: 'EQ', value: '"US"' }),
  })
  if (!r.ok) throw new Error(`createConstraint failed: ${r.status}`)
  return r.json() as Promise<Constraint>
}

export async function createVariant(flagId: number, key: string): Promise<Variant> {
  const r = await fetch(`${API}/flags/${flagId}/variants`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ key }),
  })
  if (!r.ok) throw new Error(`createVariant(${key}) failed: ${r.status}`)
  return r.json() as Promise<Variant>
}

function sleep(ms: number): Promise<void> {
  const { promise, resolve } = Promise.withResolvers<void>()
  setTimeout(resolve, ms)
  return promise
}

export async function waitForSnapshot(
  flagId: number,
  { timeout = 3000 }: { timeout?: number } = {},
): Promise<FlagSnapshot[]> {
  const deadline = Date.now() + timeout
  while (Date.now() < deadline) {
    const r = await fetch(`${API}/flags/${flagId}/snapshots`)
    if (!r.ok) throw new Error(`fetchSnapshots(${flagId}) failed: ${r.status}`)
    const snaps = (await r.json()) as FlagSnapshot[]
    if (Array.isArray(snaps) && snaps.length > 0) return snaps
    await sleep(200)
  }
  throw new Error(`waitForSnapshot(${flagId}) timed out after ${timeout}ms`)
}

export async function getSnapshotMaxId(): Promise<number> {
  const r = await fetch(`${API}/flags/snapshots/max_id`)
  if (!r.ok) throw new Error(`getSnapshotMaxId failed: ${r.status}`)
  const data = (await r.json()) as { maxID: number }
  return data.maxID
}