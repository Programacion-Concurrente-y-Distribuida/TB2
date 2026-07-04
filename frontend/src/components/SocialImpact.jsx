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
  const atRisk = data.levels
    .filter((l) => LEVEL_RISK[l.level])
    .reduce((s, l) => s + l.population, 0)

  const total = data.levels.reduce((s, l) => s + l.population, 0)
  const pct = total > 0 ? Math.round((atRisk / total) * 100) : 0

  const sorted = [...data.levels].sort(
    (a, b) => LEVEL_ORDER.indexOf(a.level) - LEVEL_ORDER.indexOf(b.level),
  )

  return (
    <div className="impact-card">
      <h4 className="impact-card__title">{POLLUTANT_LABELS[data.pollutant] || data.pollutant}</h4>
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
                <div
                  className="impact-bar__fill"
                  style={{ width: `${barPct}%`, backgroundColor: l.color }}
                />
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
    return <div className="impact-error">No se pudo cargar las estadísticas.</div>
  }

  return (
    <section className="social-impact">
      <div className="social-impact__header">
        <h2>Impacto Social</h2>
        <p>
          Estimación de la población de Lima Metropolitana expuesta a diferentes niveles de
          contaminación del aire según el modelo predictivo entrenado con datos AQS-EPA.
        </p>
      </div>

      <div className="impact-summary">
        <div className="impact-summary__stat">
          <span className="impact-summary__value">{stats.totalDistricts}</span>
          <span className="impact-summary__label">Distritos monitoreados</span>
        </div>
        <div className="impact-summary__stat">
          <span className="impact-summary__value">{formatPop(stats.totalPopulation)}</span>
          <span className="impact-summary__label">Población estimada</span>
        </div>
        <div className="impact-summary__stat">
          <span className="impact-summary__value">{stats.modelAvailable ? 'Activo' : 'Sin modelo'}</span>
          <span className="impact-summary__label">Estado del sistema</span>
        </div>
      </div>

      {!stats.modelAvailable && (
        <div className="impact-no-model">
          Aún no hay un modelo entrenado. Las predicciones estarán disponibles una vez que
          un administrador inicie el entrenamiento distribuido.
        </div>
      )}

      {stats.modelAvailable && stats.byPollutant && (
        <div className="impact-cards">
          {stats.byPollutant.map((d) => (
            <PollutantCard key={d.pollutant} data={d} />
          ))}
        </div>
      )}

      <div className="impact-disclaimer">
        <strong>Nota:</strong> Este sistema responde la pregunta: <em>"¿Cuándo se disparará la
        contaminación del aire en Lima para alertar a personas con enfermedades respiratorias?"</em>.
        Los datos de entrenamiento provienen de estaciones EPA-AQS de EE.UU. y se adaptan a las
        coordenadas de Lima como demostración académica de un sistema distribuido de ML.
      </div>
    </section>
  )
}
