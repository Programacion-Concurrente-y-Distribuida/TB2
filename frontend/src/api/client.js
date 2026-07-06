const API_BASE = import.meta.env.VITE_API_BASE_URL || ''

// ── helpers ──────────────────────────────────────────────────────────────────

async function request(path, options = {}) {
  const res = await fetch(`${API_BASE}${path}`, options)
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    const err = new Error(body.error || `Error ${res.status}`)
    err.status = res.status
    throw err
  }
  return res.json()
}

function authHeaders(token) {
  return { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` }
}

// ── public endpoints ─────────────────────────────────────────────────────────

export function fetchDistricts() {
  return request('/api/districts')
}

export function fetchDistrictPredictions(pollutant) {
  return request(`/api/districts/predict?pollutant=${encodeURIComponent(pollutant)}`)
}

export function fetchHealth() {
  return request('/api/health')
}

export function fetchStats() {
  return request('/api/stats')
}

export function fetchDatasetInfo() {
  return request('/api/dataset/info')
}

export function listDatasets(token) {
  return request('/api/dataset/list', { headers: authHeaders(token) })
}

export function selectDataset(path, token) {
  return request('/api/dataset/select', {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({ path }),
  })
}

export function migrateDataset(token) {
  return request('/api/dataset/migrate', {
    method: 'POST',
    headers: authHeaders(token),
  })
}

export function getDatasetStatus(token) {
  return request('/api/dataset/status', { headers: authHeaders(token) })
}

export function predictPublic(districtId, pollutant) {
  return request('/api/predict/public', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ districtId, pollutant }),
  })
}

// ── auth ─────────────────────────────────────────────────────────────────────

export function login(username, password) {
  return request('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  })
}

// ── protected endpoints (require JWT token) ───────────────────────────────────

export function startTraining({ solver = 'ridge', lambda = 1.0, maxRows = 0 }, token) {
  return request('/api/train', {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({ solver, lambda, maxRows }),
  })
}

export function getTrainStatus(jobId, token) {
  return request(`/api/train/${encodeURIComponent(jobId)}`, {
    headers: authHeaders(token),
  })
}

export function getModels(token) {
  return request('/api/models', { headers: authHeaders(token) })
}

export function getClusterNodes(token) {
  return request('/api/cluster/nodes', { headers: authHeaders(token) })
}

export function setClusterNodes(nodes, token) {
  return request('/api/cluster/nodes', {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({ nodes }),
  })
}

export function resetClusterNodes(token) {
  return request('/api/cluster/nodes', {
    method: 'DELETE',
    headers: authHeaders(token),
  })
}

// ── WebSocket ─────────────────────────────────────────────────────────────────

/**
 * Opens a WebSocket connection to /ws.
 * Returns the WebSocket instance; call .close() to disconnect.
 *
 * onMessage(data) is called with the parsed JSON object for every message.
 * onOpen / onClose / onError are optional callbacks.
 */
export function connectWebSocket({ onMessage, onOpen, onClose, onError } = {}) {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const base = API_BASE
    ? API_BASE.replace(/^https?:/, proto)
    : `${proto}//${window.location.host}`
  const ws = new WebSocket(`${base}/ws`)

  ws.onopen = () => onOpen?.()
  ws.onclose = () => onClose?.()
  ws.onerror = (e) => onError?.(e)
  ws.onmessage = (e) => {
    try {
      onMessage?.(JSON.parse(e.data))
    } catch {
      // ignore malformed frames
    }
  }
  return ws
}
