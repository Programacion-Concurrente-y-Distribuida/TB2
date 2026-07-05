import { useEffect, useRef, useState } from 'react'

const TYPE_META = {
  jobUpdate:        { icon: '⚙', color: '#2563eb', label: 'Job' },
  phase:            { icon: '▶', color: '#7c3aed', label: 'Fase' },
  nodeStart:        { icon: '→', color: '#0891b2', label: 'Nodo' },
  nodeDone:         { icon: '✓', color: '#16a34a', label: 'Nodo' },
  trainingComplete: { icon: '🏁', color: '#16a34a', label: 'Completado' },
  error:            { icon: '✕', color: '#dc2626', label: 'Error' },
  migrateStart:     { icon: '⇄', color: '#d97706', label: 'Migración' },
  migrateProgress:  { icon: '↑', color: '#d97706', label: 'Migración' },
  migrateDone:      { icon: '✓', color: '#16a34a', label: 'Migración' },
  migrateError:     { icon: '✕', color: '#dc2626', label: 'Migración' },
  default:          { icon: '·', color: '#6b7280', label: 'Evento' },
}

function formatTime(d) {
  return d.toLocaleTimeString('es-PE', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function formatMessage(msg) {
  const { type } = msg
  if (type === 'jobUpdate') return `Job ${msg.jobId?.slice(0, 10)}… → ${msg.status} (${msg.progress ?? 0}%)`
  if (type === 'phase') return `Fase: ${msg.phase}`
  if (type === 'nodeStart') return `Nodo iniciado: ${msg.node}`
  if (type === 'nodeDone') return `Nodo completado: ${msg.node} (${msg.progress ?? 0}%)`
  if (type === 'trainingComplete') return `Entrenamiento completado — R²: ${msg.r2?.toFixed(4)}, MAE: ${msg.mae?.toFixed(4)}`
  if (type === 'error') return `Error: ${msg.message}`
  if (type === 'migrateStart') return `Migración iniciada: ${msg.path?.split('/').pop()}`
  if (type === 'migrateProgress') return `Migración: ${msg.inserted?.toLocaleString()} docs (${msg.progress}%)`
  if (type === 'migrateDone') return `Migración completada: ${msg.inserted?.toLocaleString()} documentos`
  if (type === 'migrateError') return `Error en migración: ${msg.message}`
  return JSON.stringify(msg)
}

export default function LogPanel({ open, onClose, messages }) {
  const bottomRef = useRef(null)
  const [filter, setFilter] = useState('all')

  useEffect(() => {
    if (open) bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, open])

  const filtered = filter === 'all'
    ? messages
    : messages.filter((m) => {
        if (filter === 'errors') return m.type === 'error' || m.type === 'migrateError'
        if (filter === 'training') return ['jobUpdate', 'phase', 'nodeStart', 'nodeDone', 'trainingComplete'].includes(m.type)
        if (filter === 'migration') return m.type.startsWith('migrate')
        return true
      })

  return (
    <>
      <div className={`log-backdrop ${open ? 'log-backdrop--visible' : ''}`} onClick={onClose} />
      <div className={`log-panel ${open ? 'log-panel--open' : ''}`}>
        <div className="log-panel__header">
          <div className="log-panel__title">
            <span className="log-panel__icon">⌨</span>
            Logs del sistema
            <span className="log-panel__count">{messages.length}</span>
          </div>
          <div className="log-panel__actions">
            <select className="log-filter" value={filter} onChange={(e) => setFilter(e.target.value)}>
              <option value="all">Todos</option>
              <option value="training">Entrenamiento</option>
              <option value="migration">Migración</option>
              <option value="errors">Errores</option>
            </select>
            <button className="log-panel__close" onClick={onClose} aria-label="Cerrar">✕</button>
          </div>
        </div>

        <div className="log-panel__body">
          {filtered.length === 0 && (
            <div className="log-empty">
              <span>Sin eventos aún</span>
            </div>
          )}
          {filtered.map((entry, i) => {
            const meta = TYPE_META[entry.type] || TYPE_META.default
            return (
              <div key={i} className="log-entry">
                <span className="log-entry__time">{formatTime(entry.ts)}</span>
                <span className="log-entry__icon" style={{ color: meta.color }}>{meta.icon}</span>
                <span className="log-entry__label" style={{ color: meta.color }}>{meta.label}</span>
                <span className="log-entry__msg">{formatMessage(entry)}</span>
              </div>
            )
          })}
          <div ref={bottomRef} />
        </div>
      </div>
    </>
  )
}
