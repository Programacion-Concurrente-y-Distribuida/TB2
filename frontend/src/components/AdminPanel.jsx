import { useCallback, useEffect, useRef, useState } from 'react'
import { getClusterNodes, getModels, startTraining } from '../api/client.js'

// ── small sub-components ──────────────────────────────────────────────────────

function NodeBadge({ node }) {
  const ok = node.status === 'ok'
  return (
    <div className={`node-badge ${ok ? 'node-badge--ok' : 'node-badge--err'}`}>
      <span className="node-badge__dot" />
      <span className="node-badge__addr">{node.addr}</span>
      {ok && <span className="node-badge__lat">{node.latencyMs} ms</span>}
      {!ok && <span className="node-badge__lat">no disponible</span>}
    </div>
  )
}

function ProgressBar({ value }) {
  return (
    <div className="progress-bar">
      <div className="progress-bar__fill" style={{ width: `${value}%` }} />
      <span className="progress-bar__label">{value}%</span>
    </div>
  )
}

function ModelRow({ model }) {
  const date = new Date(model.trainedAt).toLocaleString('es-PE')
  return (
    <tr>
      <td title={model.jobId}>{model.jobId.slice(0, 14)}…</td>
      <td>{model.solver}</td>
      <td>{model.r2?.toFixed(4)}</td>
      <td>{model.mae?.toFixed(4)}</td>
      <td>{model.rmse?.toFixed(4)}</td>
      <td>{model.trainRows?.toLocaleString()}</td>
      <td>{date}</td>
    </tr>
  )
}

// ── main panel ────────────────────────────────────────────────────────────────

export default function AdminPanel({ token, wsMessage, onClose }) {
  // Cluster state
  const [nodes, setNodes] = useState([])
  const [nodesLoading, setNodesLoading] = useState(false)

  // Models state
  const [models, setModels] = useState([])
  const [modelsLoading, setModelsLoading] = useState(false)

  // Training form state
  const [solver, setSolver] = useState('ridge')
  const [lambda, setLambda] = useState(1.0)
  const [maxRows, setMaxRows] = useState(0)
  const [training, setTraining] = useState(false)

  // Active job state
  const [job, setJob] = useState(null) // { jobId, status, progress, phases, nodes }

  const jobRef = useRef(null)

  // ── load cluster nodes ──
  const refreshNodes = useCallback(() => {
    setNodesLoading(true)
    getClusterNodes(token)
      .then(setNodes)
      .catch(() => {})
      .finally(() => setNodesLoading(false))
  }, [token])

  // ── load models list ──
  const refreshModels = useCallback(() => {
    setModelsLoading(true)
    getModels(token)
      .then(setModels)
      .catch(() => {})
      .finally(() => setModelsLoading(false))
  }, [token])

  useEffect(() => {
    refreshNodes()
    refreshModels()
  }, [refreshNodes, refreshModels])

  // ── handle WebSocket messages ──
  useEffect(() => {
    if (!wsMessage) return
    const { type, jobId, status, progress, phase, node, mae, rmse, r2, message } = wsMessage

    // Only track updates for the current job
    if (jobRef.current && jobId && jobId !== jobRef.current) return

    if (type === 'jobUpdate') {
      setJob((prev) => prev ? { ...prev, status, progress: progress ?? prev.progress } : prev)
      if (status === 'done') {
        setTraining(false)
        refreshModels()
      }
      if (status === 'failed') {
        setTraining(false)
      }
    }

    if (type === 'phase') {
      setJob((prev) => prev ? { ...prev, phase } : prev)
    }

    if (type === 'nodeStart') {
      setJob((prev) => {
        if (!prev) return prev
        const activeNodes = [...(prev.activeNodes || []), node]
        return { ...prev, activeNodes }
      })
    }

    if (type === 'nodeDone') {
      setJob((prev) => {
        if (!prev) return prev
        const doneNodes = [...(prev.doneNodes || []), node]
        return { ...prev, doneNodes, progress: progress ?? prev.progress }
      })
    }

    if (type === 'trainingComplete') {
      setJob((prev) =>
        prev ? { ...prev, status: 'done', progress: 100, mae, rmse, r2 } : prev,
      )
      setTraining(false)
      refreshModels()
    }

    if (type === 'error') {
      setJob((prev) =>
        prev ? { ...prev, status: 'failed', errorMsg: message } : prev,
      )
      setTraining(false)
    }
  }, [wsMessage, refreshModels])

  // ── start training ──
  async function handleTrain(e) {
    e.preventDefault()
    if (training) return
    setTraining(true)
    setJob(null)
    try {
      const lambdaNum = parseFloat(lambda)
      const maxRowsNum = parseInt(maxRows, 10) || 0
      const { jobId } = await startTraining(
        { solver, lambda: isNaN(lambdaNum) ? 1.0 : lambdaNum, maxRows: maxRowsNum },
        token,
      )
      jobRef.current = jobId
      setJob({ jobId, status: 'pending', progress: 0, phase: null, activeNodes: [], doneNodes: [] })
    } catch (err) {
      setTraining(false)
      setJob({ status: 'failed', errorMsg: err.message })
    }
  }

  const phaseLabel = {
    loadingData: 'Cargando datos del CSV',
    distributing: 'Distribuyendo a nodos ML',
    solving: 'Resolviendo ecuaciones normales',
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal modal--wide" onClick={(e) => e.stopPropagation()}>
        <div className="modal__header">
          <h2>Panel de Administración</h2>
          <button className="modal__close" onClick={onClose} aria-label="Cerrar">✕</button>
        </div>

        <div className="admin-panel">

          {/* ── Cluster de Nodos ── */}
          <section className="admin-section">
            <div className="admin-section__title">
              <h3>Cluster de Nodos ML</h3>
              <button className="btn btn--sm" onClick={refreshNodes} disabled={nodesLoading}>
                {nodesLoading ? '…' : 'Actualizar'}
              </button>
            </div>
            <div className="node-list">
              {nodes.length === 0 && !nodesLoading && (
                <p className="admin-empty">Sin datos de nodos</p>
              )}
              {nodes.map((n) => (
                <NodeBadge key={n.addr} node={n} />
              ))}
            </div>
          </section>

          {/* ── Iniciar Entrenamiento ── */}
          <section className="admin-section">
            <h3>Iniciar Entrenamiento Distribuido</h3>
            <form className="train-form" onSubmit={handleTrain}>
              <label className="form-label">
                Solver
                <select
                  className="form-input"
                  value={solver}
                  onChange={(e) => setSolver(e.target.value)}
                  disabled={training}
                >
                  <option value="ridge">Ridge (recomendado)</option>
                  <option value="svd">SVD</option>
                  <option value="normal">Ecuaciones normales</option>
                </select>
              </label>

              <label className="form-label">
                Lambda (regularización)
                <input
                  className="form-input"
                  type="number"
                  min="0"
                  step="0.1"
                  value={lambda}
                  onChange={(e) => setLambda(e.target.value)}
                  disabled={training}
                />
              </label>

              <label className="form-label">
                Máximo de filas (0 = todas)
                <input
                  className="form-input"
                  type="number"
                  min="0"
                  step="10000"
                  value={maxRows}
                  onChange={(e) => setMaxRows(e.target.value)}
                  disabled={training}
                />
              </label>

              <button className="btn btn--primary" type="submit" disabled={training}>
                {training ? 'Entrenando…' : 'Iniciar Entrenamiento'}
              </button>
            </form>

            {/* ── Progreso del Job ── */}
            {job && (
              <div className="job-status">
                <div className="job-status__row">
                  <span className="job-status__label">Job ID:</span>
                  <code>{job.jobId}</code>
                </div>
                <div className="job-status__row">
                  <span className="job-status__label">Estado:</span>
                  <span className={`job-badge job-badge--${job.status}`}>
                    {job.status}
                  </span>
                </div>
                {job.phase && (
                  <div className="job-status__row">
                    <span className="job-status__label">Fase:</span>
                    <span>{phaseLabel[job.phase] || job.phase}</span>
                  </div>
                )}

                <ProgressBar value={job.progress || 0} />

                {job.doneNodes?.length > 0 && (
                  <div className="job-status__row">
                    <span className="job-status__label">Nodos completados:</span>
                    <span>{job.doneNodes.join(', ')}</span>
                  </div>
                )}

                {job.status === 'done' && (
                  <div className="job-metrics">
                    <div className="metric">
                      <span className="metric__label">R²</span>
                      <span className="metric__value">{job.r2?.toFixed(4)}</span>
                    </div>
                    <div className="metric">
                      <span className="metric__label">MAE</span>
                      <span className="metric__value">{job.mae?.toFixed(4)}</span>
                    </div>
                    <div className="metric">
                      <span className="metric__label">RMSE</span>
                      <span className="metric__value">{job.rmse?.toFixed(4)}</span>
                    </div>
                  </div>
                )}

                {job.errorMsg && (
                  <p className="job-status__error">{job.errorMsg}</p>
                )}
              </div>
            )}
          </section>

          {/* ── Lista de Modelos ── */}
          <section className="admin-section">
            <div className="admin-section__title">
              <h3>Modelos Entrenados</h3>
              <button className="btn btn--sm" onClick={refreshModels} disabled={modelsLoading}>
                {modelsLoading ? '…' : 'Actualizar'}
              </button>
            </div>
            {models.length === 0 && !modelsLoading && (
              <p className="admin-empty">
                No hay modelos aún. Inicia un entrenamiento para crear el primero.
              </p>
            )}
            {models.length > 0 && (
              <div className="table-scroll">
                <table className="models-table">
                  <thead>
                    <tr>
                      <th>Job ID</th>
                      <th>Solver</th>
                      <th>R²</th>
                      <th>MAE</th>
                      <th>RMSE</th>
                      <th>Filas train</th>
                      <th>Entrenado</th>
                    </tr>
                  </thead>
                  <tbody>
                    {models.map((m) => (
                      <ModelRow key={m.jobId || m._id} model={m} />
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </section>
        </div>
      </div>
    </div>
  )
}
