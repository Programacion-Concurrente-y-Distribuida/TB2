import { useState } from 'react'
import { POLLUTANTS } from '../utils/aqi.js'

const API_BASE = import.meta.env.VITE_API_BASE_URL || ''

const DISTRICTS = [
  { id: '150101', name: 'Lima (Cercado)' },
  { id: '150102', name: 'Ancón' },
  { id: '150103', name: 'Ate' },
  { id: '150104', name: 'Barranco' },
  { id: '150105', name: 'Breña' },
  { id: '150106', name: 'Carabayllo' },
  { id: '150107', name: 'Chaclacayo' },
  { id: '150108', name: 'Chorrillos' },
  { id: '150109', name: 'Cieneguilla' },
  { id: '150110', name: 'Comas' },
  { id: '150111', name: 'El Agustino' },
  { id: '150112', name: 'Independencia' },
  { id: '150113', name: 'Jesús María' },
  { id: '150114', name: 'La Molina' },
  { id: '150115', name: 'La Victoria' },
  { id: '150116', name: 'Lince' },
  { id: '150117', name: 'Los Olivos' },
  { id: '150118', name: 'Lurigancho' },
  { id: '150119', name: 'Lurín' },
  { id: '150120', name: 'Magdalena del Mar' },
  { id: '150121', name: 'Pueblo Libre' },
  { id: '150122', name: 'Miraflores' },
  { id: '150123', name: 'Pachacámac' },
  { id: '150124', name: 'Pucusana' },
  { id: '150125', name: 'Puente Piedra' },
  { id: '150126', name: 'Punta Hermosa' },
  { id: '150127', name: 'Punta Negra' },
  { id: '150128', name: 'Rímac' },
  { id: '150129', name: 'San Bartolo' },
  { id: '150130', name: 'San Borja' },
  { id: '150131', name: 'San Isidro' },
  { id: '150132', name: 'San Juan de Lurigancho' },
  { id: '150133', name: 'San Juan de Miraflores' },
  { id: '150134', name: 'San Luis' },
  { id: '150135', name: 'San Martín de Porres' },
  { id: '150136', name: 'San Miguel' },
  { id: '150137', name: 'Santa Anita' },
  { id: '150138', name: 'Santa María del Mar' },
  { id: '150139', name: 'Santa Rosa' },
  { id: '150140', name: 'Santiago de Surco' },
  { id: '150141', name: 'Surquillo' },
  { id: '150142', name: 'Villa El Salvador' },
  { id: '150143', name: 'Villa María del Triunfo' },
]

const LEVEL_ADVICE = {
  'Bueno': 'La calidad del aire es satisfactoria. Actividades al aire libre son seguras.',
  'Moderado': 'Calidad aceptable. Personas muy sensibles deben limitar esfuerzo prolongado.',
  'Dañino (grupos sensibles)': 'Niños, ancianos y personas con enfermedades respiratorias deben reducir actividad exterior.',
  'Dañino': 'Toda la población puede comenzar a sentir efectos. Evitar actividad prolongada al aire libre.',
  'Muy dañino': 'Alerta de salud. Toda la población debe evitar actividad al aire libre.',
  'Peligroso': 'Emergencia sanitaria. Permanecer en interiores. Usar mascarilla N95 si debe salir.',
}

export default function PredictPanel() {
  const [districtId, setDistrictId] = useState('150101')
  const [pollutant, setPollutant] = useState('PM2.5')
  const [result, setResult] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function handleSubmit(e) {
    e.preventDefault()
    setLoading(true)
    setError('')
    setResult(null)
    try {
      const res = await fetch(`${API_BASE}/api/predict/public`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ districtId, pollutant }),
      })
      const data = await res.json()
      if (!res.ok) throw new Error(data.error || `Error ${res.status}`)
      setResult(data)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="predict-panel">
      <h2 className="predict-panel__title">Consulta de Calidad del Aire</h2>
      <p className="predict-panel__subtitle">
        Obtén una predicción de contaminación para cualquier distrito de Lima Metropolitana.
      </p>

      <form className="predict-form" onSubmit={handleSubmit}>
        <div className="predict-form__row">
          <label className="form-label">
            Distrito
            <select
              className="form-input"
              value={districtId}
              onChange={(e) => setDistrictId(e.target.value)}
              disabled={loading}
            >
              {DISTRICTS.map((d) => (
                <option key={d.id} value={d.id}>{d.name}</option>
              ))}
            </select>
          </label>

          <label className="form-label">
            Contaminante
            <select
              className="form-input"
              value={pollutant}
              onChange={(e) => setPollutant(e.target.value)}
              disabled={loading}
            >
              {POLLUTANTS.map((p) => (
                <option key={p.id} value={p.id}>{p.label} ({p.unit})</option>
              ))}
            </select>
          </label>
        </div>

        <button className="btn btn--primary" type="submit" disabled={loading}>
          {loading ? 'Calculando…' : 'Predecir'}
        </button>
      </form>

      {error && <p className="predict-panel__error">{error}</p>}

      {result && (
        <div className="predict-result" style={{ borderLeftColor: result.color }}>
          <div className="predict-result__header">
            <span className="predict-result__district">{result.district}</span>
            <span className="predict-result__level" style={{ backgroundColor: result.color }}>
              {result.level}
            </span>
          </div>
          <div className="predict-result__value">
            <span className="predict-result__number">{result.prediction.toFixed(2)}</span>
            <span className="predict-result__unit">{result.unit}</span>
          </div>
          <p className="predict-result__advice">
            {LEVEL_ADVICE[result.level] || ''}
          </p>
        </div>
      )}
    </div>
  )
}
