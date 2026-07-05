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
  'Bueno': 'La calidad del aire es satisfactoria. Actividades al aire libre son completamente seguras para toda la población.',
  'Moderado': 'Calidad aceptable. Personas muy sensibles a la contaminación deben limitar el esfuerzo prolongado al exterior.',
  'Dañino (grupos sensibles)': 'Niños, adultos mayores y personas con enfermedades respiratorias o cardíacas deben reducir la actividad exterior.',
  'Dañino': 'Toda la población puede comenzar a sentir efectos adversos. Evitar actividad prolongada al aire libre.',
  'Muy dañino': 'Alerta de salud. Toda la población debe evitar actividad al aire libre. Vulnerable debe permanecer en interiores.',
  'Peligroso': 'Emergencia sanitaria. Permanecer en interiores con ventanas cerradas. Usar mascarilla N95 si debe salir.',
}

const AQI_LEVELS = ['Bueno', 'Moderado', 'Dañino (grupos sensibles)', 'Dañino', 'Muy dañino', 'Peligroso']
const AQI_COLORS = ['#22c55e', '#eab308', '#f97316', '#ef4444', '#a855f7', '#7f1d1d']

export default function PredictPanel({ toast }) {
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
      toast?.error(err.message)
    } finally {
      setLoading(false)
    }
  }

  const levelIdx = result ? AQI_LEVELS.indexOf(result.level) : -1

  return (
    <div className="predict-panel">
      <div className="predict-panel__heading">
        <h2>Consulta de Calidad del Aire</h2>
        <p>Predicción de contaminación por distrito de Lima Metropolitana usando el modelo distribuido entrenado.</p>
      </div>

      <div className="predict-layout">
        <div className="predict-form-card">
          <h3>Parámetros de consulta</h3>

          <form onSubmit={handleSubmit}>
            <label className="form-label" style={{ marginBottom: '0.75rem' }}>
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

            <label className="form-label" style={{ marginBottom: '1rem' }}>
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

            <button className="btn btn--primary btn--lg" type="submit" disabled={loading} style={{ width: '100%' }}>
              {loading ? '⟳ Calculando predicción…' : '🔍 Predecir calidad del aire'}
            </button>
          </form>

          {error && (
            <div className="predict-error">
              <span>⚠</span> {error}
            </div>
          )}
        </div>

        <div>
          {result ? (
            <div className="predict-result-card">
              <div className="predict-result-card__top" style={{ background: `linear-gradient(135deg, ${result.color}dd, ${result.color})` }}>
                <div className="predict-result-card__district">{result.district}</div>
                <div className="predict-result-card__value">{result.prediction.toFixed(2)}</div>
                <div className="predict-result-card__unit">{result.unit}</div>
                <div className="predict-result-card__level">{result.level}</div>

                <div className="aqi-gauge" style={{ marginTop: '1rem' }}>
                  {AQI_LEVELS.map((lvl, i) => (
                    <div
                      key={lvl}
                      className={`aqi-gauge__level ${i <= levelIdx ? 'aqi-gauge__level--active' : ''}`}
                      style={{ background: AQI_COLORS[i] }}
                      title={lvl}
                    />
                  ))}
                </div>
              </div>
              <div className="predict-result-card__bottom">
                <p className="predict-result-card__advice">
                  {LEVEL_ADVICE[result.level] || ''}
                </p>
              </div>
            </div>
          ) : (
            <div className="predict-empty">
              <div className="predict-empty__icon">🌬</div>
              <p>Selecciona un distrito y contaminante, luego presiona <strong>Predecir</strong> para obtener el resultado del modelo ML.</p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
