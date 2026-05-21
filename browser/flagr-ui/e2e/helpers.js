const API = process.env.API_URL || 'http://localhost:18000/api/v1'

export async function createFlag(opts = {}) {
  const r = await fetch(`${API}/flags`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ description: opts.description || `e2e-${Date.now()}` })
  })
  const flag = await r.json()
  return flag
}

export async function createFlagWithVariants(opts = {}) {
  const flag = await createFlag(opts)
  await fetch(`${API}/flags/${flag.id}/variants`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ key: 'control' })
  })
  await fetch(`${API}/flags/${flag.id}/variants`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ key: 'test' })
  })
  return flag
}

export async function deleteFlag(flagId) {
  try { await fetch(`${API}/flags/${flagId}`, { method: 'DELETE' }) } catch {}
}

export async function createSegment(flagId, desc) {
  const r = await fetch(`${API}/flags/${flagId}/segments`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ description: desc, rolloutPercent: 100 })
  })
  return r.json()
}

export async function createConstraint(flagId, segId) {
  const r = await fetch(`${API}/flags/${flagId}/segments/${segId}/constraints`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ property: 'country', operator: 'EQ', value: '"US"' })
  })
  return r.json()
}

export async function createVariant(flagId, key) {
  const r = await fetch(`${API}/flags/${flagId}/variants`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ key })
  })
  return r.json()
}
