const API = 'http://127.0.0.1:18000/api/v1'

async function createFlag(description) {
  const res = await fetch(`${API}/flags`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ description: description || 'test-flag-' + Date.now() }),
  })
  return res.json()
}

async function createVariant(flagId, key) {
  const res = await fetch(`${API}/flags/${flagId}/variants`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ key }),
  })
  return res.json()
}

async function createSegment(flagId, description, rolloutPercent = 50) {
  const res = await fetch(`${API}/flags/${flagId}/segments`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ description, rolloutPercent }),
  })
  return res.json()
}

module.exports = { API, createFlag, createVariant, createSegment }
