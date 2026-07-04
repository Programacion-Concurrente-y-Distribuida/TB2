const API_BASE = import.meta.env.VITE_API_BASE_URL || ''

async function request(path) {
  const res = await fetch(`${API_BASE}${path}`)
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `Error ${res.status}`)
  }
  return res.json()
}

export function fetchDistricts() {
  return request('/api/districts')
}

export function fetchDistrictPredictions(pollutant) {
  return request(`/api/districts/predict?pollutant=${encodeURIComponent(pollutant)}`)
}
