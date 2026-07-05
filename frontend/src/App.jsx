import { useEffect, useState } from 'react'
import MapView from './components/MapView.jsx'
import PollutantSelector from './components/PollutantSelector.jsx'
import Legend from './components/Legend.jsx'
import DistrictPanel from './components/DistrictPanel.jsx'
import LoginModal from './components/LoginModal.jsx'
import AdminPanel from './components/AdminPanel.jsx'
import PredictPanel from './components/PredictPanel.jsx'
import SocialImpact from './components/SocialImpact.jsx'
import ToastContainer from './components/Toast.jsx'
import LogPanel from './components/LogPanel.jsx'
import { fetchDistrictPredictions } from './api/client.js'
import { useWebSocket } from './hooks/useWebSocket.js'
import { useToast } from './hooks/useToast.js'
import './App.css'

const TOKEN_KEY = 'aqsml_token'

function loadToken() {
  try { return localStorage.getItem(TOKEN_KEY) } catch { return null }
}
function saveToken(t) {
  try { localStorage.setItem(TOKEN_KEY, t) } catch {}
}
function clearToken() {
  try { localStorage.removeItem(TOKEN_KEY) } catch {}
}

const TABS = [
  { id: 'map', label: '🗺 Mapa' },
  { id: 'predict', label: '🔍 Predicción' },
  { id: 'impact', label: '📊 Impacto Social' },
]

export default function App() {
  const [activeTab, setActiveTab] = useState('map')
  const [pollutant, setPollutant] = useState('PM2.5')
  const [predictions, setPredictions] = useState([])
  const [selected, setSelected] = useState(null)
  const [status, setStatus] = useState('loading')

  const [token, setToken] = useState(loadToken)
  const [showLogin, setShowLogin] = useState(false)
  const [showAdmin, setShowAdmin] = useState(false)
  const [showLogs, setShowLogs] = useState(false)
  const [logMessages, setLogMessages] = useState([])
  const [errorCount, setErrorCount] = useState(0)

  const { connected: wsConnected, lastMessage } = useWebSocket()
  const { toasts, removeToast, toast } = useToast()

  useEffect(() => {
    if (activeTab !== 'map') return
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
    return () => { cancelled = true }
  }, [pollutant, activeTab])

  useEffect(() => {
    if (!lastMessage) return

    const entry = { ...lastMessage, ts: new Date() }
    setLogMessages((prev) => [...prev.slice(-499), entry])

    if (lastMessage.type === 'error' || lastMessage.type === 'migrateError') {
      setErrorCount((n) => n + 1)
      toast.error(lastMessage.message || 'Error desconocido')
    }

    if (lastMessage.type === 'trainingComplete') {
      toast.success(`Entrenamiento completado — R²: ${lastMessage.r2?.toFixed(4)}`)
      const t = setTimeout(() => {
        setStatus('loading')
        fetchDistrictPredictions(pollutant)
          .then((data) => {
            setPredictions(data.districts || [])
            setSelected(null)
            setStatus('ready')
          })
          .catch(() => setStatus('error'))
      }, 1500)
      return () => clearTimeout(t)
    }

    if (lastMessage.type === 'migrateDone') {
      toast.success(`Migración completada — ${lastMessage.inserted?.toLocaleString()} docs`)
    }
  }, [lastMessage, pollutant])

  function handleLogin(newToken) {
    saveToken(newToken)
    setToken(newToken)
    setShowLogin(false)
    setShowAdmin(true)
    toast.success('Sesión iniciada correctamente')
  }

  function handleLogout() {
    clearToken()
    setToken(null)
    setShowAdmin(false)
    toast.info('Sesión cerrada')
  }

  function handleOpenLogs() {
    setErrorCount(0)
    setShowLogs(true)
  }

  return (
    <div className="app">
      <header className="app__header">
        <div className="app__header-brand">
          <span className="app__header-brand-icon">🌫</span>
          <div className="app__header-left">
            <h1>Calidad del Aire — Lima</h1>
            <p>Sistema distribuido de predicción de contaminación por distrito</p>
          </div>
        </div>

        <div className="app__header-right">
          <span className={`ws-indicator ${wsConnected ? 'ws-indicator--on' : 'ws-indicator--off'}`}>
            {wsConnected ? 'En vivo' : 'Desconectado'}
          </span>
          <button className="log-toggle" onClick={handleOpenLogs} title="Ver logs del sistema">
            ⌨ Logs
            {errorCount > 0 && <span className="log-toggle__badge">{errorCount}</span>}
          </button>
          {token ? (
            <>
              <button className="btn btn--sm btn--outline" onClick={() => setShowAdmin(true)}>
                Admin
              </button>
              <button className="btn btn--sm btn--danger" onClick={handleLogout}>
                Salir
              </button>
            </>
          ) : (
            <button className="btn btn--sm btn--outline" onClick={() => setShowLogin(true)}>
              Iniciar sesión
            </button>
          )}
        </div>
      </header>

      <nav className="app__tabs">
        {TABS.map((tab) => (
          <button
            key={tab.id}
            className={`tab-btn ${activeTab === tab.id ? 'tab-btn--active' : ''}`}
            onClick={() => setActiveTab(tab.id)}
          >
            {tab.label}
          </button>
        ))}
      </nav>

      {activeTab === 'map' && (
        <div className="app__body">
          <aside className="app__sidebar">
            <PollutantSelector value={pollutant} onChange={setPollutant} />
            <Legend />
            <DistrictPanel district={selected} />
            {!token && (
              <div className="sidebar-hint">
                <p>
                  ¿Necesitas entrenar el modelo?{' '}
                  <button className="link-btn" onClick={() => setShowLogin(true)}>
                    Inicia sesión
                  </button>{' '}
                  como administrador.
                </p>
              </div>
            )}
          </aside>

          <main className="app__map">
            {status === 'loading' && (
              <div className="app__banner">Cargando predicciones…</div>
            )}
            {status === 'no-model' && (
              <div className="app__banner app__banner--warning">
                Aún no hay un modelo entrenado.{' '}
                {token ? (
                  <button className="link-btn" onClick={() => setShowAdmin(true)}>
                    Inicia un entrenamiento
                  </button>
                ) : (
                  <>
                    <button className="link-btn" onClick={() => setShowLogin(true)}>
                      Inicia sesión
                    </button>{' '}
                    como administrador y entrena el modelo.
                  </>
                )}
              </div>
            )}
            {status === 'error' && (
              <div className="app__banner app__banner--error">
                No se pudo conectar con el API. Verifica que el backend esté corriendo.
              </div>
            )}
            <MapView
              predictions={predictions}
              selectedId={selected?.id}
              onSelectDistrict={setSelected}
            />
          </main>
        </div>
      )}

      {activeTab === 'predict' && (
        <div className="app__content">
          <PredictPanel toast={toast} />
        </div>
      )}

      {activeTab === 'impact' && (
        <div className="app__content">
          <SocialImpact />
        </div>
      )}

      <footer className="app__footer">
        <span>Sistema distribuido de ML — Go + goroutines + TCP</span>
        <div className="app__footer-sources">
          <span className="footer-sources__label">Fuentes:</span>
          <a href="https://www.datosabiertos.gob.pe/dataset/monitoreo-de-los-contaminantes-del-aire-en-lima-metropolitana-servicio-nacional-de" target="_blank" rel="noopener noreferrer">SENAMHI REMCA</a>
          <span className="footer-sources__sep">·</span>
          <a href="https://www.epa.gov/aqs" target="_blank" rel="noopener noreferrer">EPA AQS</a>
          <span className="footer-sources__sep">·</span>
          <a href="https://data.humdata.org/dataset/peru-administrative-level-0-1-2-3-boundaries" target="_blank" rel="noopener noreferrer">IGN / OCHA HDX</a>
        </div>
      </footer>

      {showLogin && (
        <LoginModal onLogin={handleLogin} onClose={() => setShowLogin(false)} />
      )}

      {showAdmin && token && (
        <AdminPanel
          token={token}
          wsMessage={lastMessage}
          onClose={() => setShowAdmin(false)}
          toast={toast}
        />
      )}

      <LogPanel open={showLogs} onClose={() => setShowLogs(false)} messages={logMessages} />
      <ToastContainer toasts={toasts} onRemove={removeToast} />
    </div>
  )
}
