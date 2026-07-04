import { useEffect, useState } from 'react'
import MapView from './components/MapView.jsx'
import PollutantSelector from './components/PollutantSelector.jsx'
import Legend from './components/Legend.jsx'
import DistrictPanel from './components/DistrictPanel.jsx'
import { fetchDistrictPredictions } from './api/client.js'
import './App.css'

export default function App() {
  const [pollutant, setPollutant] = useState('PM2.5')
  const [predictions, setPredictions] = useState([])
  const [selected, setSelected] = useState(null)
  const [status, setStatus] = useState('loading') // loading | ready | error | no-model

  useEffect(() => {
    let cancelled = false
    setStatus('loading')
    fetchDistrictPredictions(pollutant)
      .then((data) => {
        if (cancelled) return
        setPredictions(data.districts || [])
        setSelected(null)
        setStatus('ready')
      })
      .catch((err) => {
        if (cancelled) return
        console.error(err)
        setStatus(/entrenado/i.test(err.message) ? 'no-model' : 'error')
      })
    return () => {
      cancelled = true
    }
  }, [pollutant])

  return (
    <div className="app">
      <header className="app__header">
        <h1>Calidad del Aire — Lima</h1>
        <p>Prediccion de contaminacion del aire por distrito (demo academica)</p>
      </header>

      <div className="app__body">
        <aside className="app__sidebar">
          <PollutantSelector value={pollutant} onChange={setPollutant} />
          <Legend />
          <DistrictPanel district={selected} />
        </aside>

        <main className="app__map">
          {status === 'no-model' && (
            <div className="app__banner">
              Aun no hay un modelo entrenado en el backend. Ejecuta un entrenamiento
              (POST /api/train) y vuelve a intentarlo.
            </div>
          )}
          {status === 'error' && (
            <div className="app__banner app__banner--error">
              No se pudo conectar con el API. Verifica que el backend este corriendo.
            </div>
          )}
          <MapView predictions={predictions} selectedId={selected?.id} onSelectDistrict={setSelected} />
        </main>
      </div>

      <footer className="app__footer">
        Limites distritales: Instituto Geográfico Nacional (IGN) vía OCHA / Humanitarian Data Exchange
        (CC BY-IGO). Modelo predictivo entrenado sobre datos de estaciones EPA AQS (EE.UU.) — ver README
        para el detalle de las simplificaciones asumidas para Lima.
      </footer>
    </div>
  )
}
