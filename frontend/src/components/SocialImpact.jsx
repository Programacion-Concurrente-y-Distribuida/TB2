import { useEffect, useState } from 'react'

const API_BASE = import.meta.env.VITE_API_BASE_URL || ''

const POLLUTANT_LABELS = {
  'PM2.5': 'Material Particulado PM2.5',
  'O3': 'Ozono (O₃)',
  'NO2': 'Dióxido de Nitrógeno (NO₂)',
  'CO': 'Monóxido de Carbono (CO)',
}

const LEVEL_ORDER = [
  'Bueno', 'Moderado', 'Dañino (grupos sensibles)', 'Dañino', 'Muy dañino', 'Peligroso',
]

const LEVEL_RISK = {
  'Bueno': false,
  'Moderado': false,
  'Dañino (grupos sensibles)': true,
  'Dañino': true,
  'Muy dañino': true,
  'Peligroso': true,
}

function formatPop(n) {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(2)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(0)}k`
  return String(n)
}

function PollutantCard({ data }) {
  const atRisk = data.levels.filter((l) => LEVEL_RISK[l.level]).reduce((s, l) => s + l.population, 0)
  const total = data.levels.reduce((s, l) => s + l.population, 0)
  const pct = total > 0 ? Math.round((atRisk / total) * 100) : 0
  const sorted = [...data.levels].sort((a, b) => LEVEL_ORDER.indexOf(a.level) - LEVEL_ORDER.indexOf(b.level))

  return (
    <div className="impact-card">
      <p className="impact-card__title">{POLLUTANT_LABELS[data.pollutant] || data.pollutant}</p>
      <div className="impact-card__risk">
        <span className="impact-card__risk-pop">{formatPop(atRisk)}</span>
        <span className="impact-card__risk-label">personas en riesgo ({pct}%)</span>
      </div>
      <div className="impact-bars">
        {sorted.map((l) => {
          const barPct = total > 0 ? Math.round((l.population / total) * 100) : 0
          return (
            <div key={l.level} className="impact-bar">
              <div className="impact-bar__label">
                <span className="impact-bar__dot" style={{ backgroundColor: l.color }} />
                <span>{l.level}</span>
              </div>
              <div className="impact-bar__track">
                <div className="impact-bar__fill" style={{ width: `${barPct}%`, backgroundColor: l.color }} />
              </div>
              <span className="impact-bar__pct">{barPct}%</span>
              <span className="impact-bar__count">{formatPop(l.population)}</span>
            </div>
          )
        })}
      </div>
    </div>
  )
}

export default function SocialImpact() {
  const [stats, setStats] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch(`${API_BASE}/api/stats`)
      .then((r) => r.json())
      .then(setStats)
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return <div className="impact-loading">Calculando impacto social…</div>
  }

  if (!stats) {
    return (
      <div className="impact-page">
        <div className="predict-empty">
          <div className="predict-empty__icon">📊</div>
          <p>No se pudo cargar las estadísticas. Verifica que el backend esté corriendo.</p>
        </div>
      </div>
    )
  }

  return (
    <div className="impact-page">
      <div className="impact-page__heading">
        <h2>Impacto Social en Lima Metropolitana</h2>
        <p>
          Estimación de la población expuesta a diferentes niveles de contaminación del aire
          según el modelo predictivo distribuido entrenado con datos reales.
        </p>
      </div>

      <div style={{ display: 'flex', gap: '1rem', marginBottom: '1.5rem', flexWrap: 'wrap' }}>
        <div className="dataset-stat" style={{ minWidth: 140 }}>
          <span className="dataset-stat__label">Distritos</span>
          <span className="dataset-stat__value">{stats.totalDistricts}</span>
        </div>
        <div className="dataset-stat" style={{ minWidth: 140 }}>
          <span className="dataset-stat__label">Población estimada</span>
          <span className="dataset-stat__value">{formatPop(stats.totalPopulation)}</span>
        </div>
        <div className="dataset-stat" style={{ minWidth: 140 }}>
          <span className="dataset-stat__label">Estado del sistema</span>
          <span className="dataset-stat__value" style={{ color: stats.modelAvailable ? 'var(--color-success)' : 'var(--color-text-muted)' }}>
            {stats.modelAvailable ? '✓ Modelo activo' : 'Sin modelo'}
          </span>
        </div>
      </div>

      {!stats.modelAvailable && (
        <div className="app__banner app__banner--warning" style={{ position: 'static', transform: 'none', maxWidth: 'none', marginBottom: '1.5rem' }}>
          ⚠ Aún no hay un modelo entrenado. Las predicciones estarán disponibles una vez que un administrador inicie el entrenamiento distribuido.
        </div>
      )}

      {stats.modelAvailable && stats.byPollutant && (
        <div className="impact-grid">
          {stats.byPollutant.map((d) => (
            <PollutantCard key={d.pollutant} data={d} />
          ))}
        </div>
      )}

      <div style={{ marginTop: '1.5rem', padding: '0.85rem 1rem', background: 'var(--color-bg)', border: '1px solid var(--color-border)', borderRadius: 'var(--radius-md)', fontSize: '0.78rem', color: 'var(--color-text-muted)', lineHeight: '1.6' }}>
        <strong>Nota metodológica:</strong> Este sistema responde la pregunta: <em>"¿Cuándo se disparará la contaminación del aire en Lima para alertar a personas con enfermedades respiratorias?"</em>. Los datos de entrenamiento provienen de estaciones SENAMHI-REMCA y se usan para demostrar un sistema distribuido de ML con particionamiento de la matriz de diseño entre múltiples nodos.
      </div>
    </div>
  )
}
