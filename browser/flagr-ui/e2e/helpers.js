export const API = process.env.API_URL || 'http://localhost:18000/api/v1'

export async function createFlag(opts = {}) {
  const r = await fetch(`${API}/flags`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ description: opts.description || `e2e-${Date.now()}` })
  })
  if (!r.ok) throw new Error(`createFlag failed: ${r.status} ${await r.text()}`)
  return r.json()
}

export async function createFlagWithVariants(opts = {}) {
  const flag = await createFlag(opts)
  for (const key of ['control', 'test']) {
    const r = await fetch(`${API}/flags/${flag.id}/variants`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ key })
    })
    if (!r.ok) throw new Error(`createVariant(${key}) failed: ${r.status}`)
  }
  return flag
}

export async function deleteFlag(flagId) {
  const r = await fetch(`${API}/flags/${flagId}`, { method: 'DELETE' })
  if (!r.ok) throw new Error(`deleteFlag(${flagId}) failed: ${r.status}`)
}

export async function createSegment(flagId, desc) {
  const r = await fetch(`${API}/flags/${flagId}/segments`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ description: desc, rolloutPercent: 100 })
  })
  if (!r.ok) throw new Error(`createSegment failed: ${r.status}`)
  return r.json()
}

export async function createConstraint(flagId, segId) {
  const r = await fetch(`${API}/flags/${flagId}/segments/${segId}/constraints`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ property: 'country', operator: 'EQ', value: '"US"' })
  })
  if (!r.ok) throw new Error(`createConstraint failed: ${r.status}`)
  return r.json()
}

export async function createVariant(flagId, key) {
  const r = await fetch(`${API}/flags/${flagId}/variants`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ key })
  })
  if (!r.ok) throw new Error(`createVariant(${key}) failed: ${r.status}`)
  return r.json()
}

/** Poll /snapshots until at least one snapshot exists, up to `timeout` ms. */
export async function waitForSnapshot(flagId, { timeout = 3000 } = {}) {
  const deadline = Date.now() + timeout
  while (Date.now() < deadline) {
    const r = await fetch(`${API}/flags/${flagId}/snapshots`)
    if (!r.ok) throw new Error(`fetchSnapshots(${flagId}) failed: ${r.status}`)
    const snaps = await r.json()
    if (Array.isArray(snaps) && snaps.length > 0) return snaps
    await sleep(200)
  }
  throw new Error(`waitForSnapshot(${flagId}) timed out after ${timeout}ms`)
}

function sleep(ms) {
  return new Promise(r => setTimeout(r, ms))
}
