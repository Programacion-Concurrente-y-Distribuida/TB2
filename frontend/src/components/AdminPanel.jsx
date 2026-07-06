import { useCallback, useEffect, useRef, useState } from 'react'
import { getClusterNodes, getModels, startTraining, setClusterNodes, resetClusterNodes, fetchDatasetInfo, listDatasets, selectDataset, migrateDataset, getDatasetStatus } from '../api/client.js'

function NodeBadge({ node, onRemove }) {
  const ok = node.status === 'ok'
  return (
    <div className={`node-badge ${ok ? 'node-badge--ok' : 'node-badge--err'}`}>
      <span className="node-badge__dot" />
      <span className="node-badge__addr">
        {node.addr}
        {node.isDefault && <span className="node-badge__tag">Docker</span>}
      </span>
      {ok && <span className="node-badge__lat">{node.latencyMs} ms</span>}
      {!ok && <span className="node-badge__lat">no disponible</span>}
      {onRemove && (
        <button className="node-badge__remove" onClick={() => onRemove(node.addr)} title="Quitar nodo">✕</button>
      )}
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

export default function AdminPanel({ token, wsMessage, onClose, toast }) {
  const [clusterData, setClusterData] = useState({ nodes: [], usingDynamic: false, defaults: [] })
  const [nodesLoading, setNodesLoading] = useState(false)
  const [nodesAuthError, setNodesAuthError] = useState(false)
  const [newNodeAddr, setNewNodeAddr] = useState('')
  const [nodeError, setNodeError] = useState('')
  const [nodeSaving, setNodeSaving] = useState(false)

  const [models, setModels] = useState([])
  const [modelsLoading, setModelsLoading] = useState(false)

  const [solver, setSolver] = useState('ridge')
  const [lambda, setLambda] = useState(1.0)
  const [maxRows, setMaxRows] = useState(0)
  const [training, setTraining] = useState(false)

  const [job, setJob] = useState(null)
  const jobRef = useRef(null)

  const [datasetInfo, setDatasetInfo] = useState(null)
  const [datasets, setDatasets] = useState([])
  const [datasetSaving, setDatasetSaving] = useState(false)
  const [mongoStatus, setMongoStatus] = useState({ ready: false, measurementsCount: 0 })
  const [migrating, setMigrating] = useState(false)
  const [migrateProgress, setMigrateProgress] = useState(null)

  const refreshNodes = useCallback(() => {
    setNodesLoading(true)
    setNodesAuthError(false)
    getClusterNodes(token)
      .then((data) => {
        if (data && Array.isArray(data.nodes)) setClusterData(data)
      })
      .catch((err) => {
        if (err.status === 401) setNodesAuthError(true)
        console.error('refreshNodes:', err)
      })
      .finally(() => setNodesLoading(false))
  }, [token])

  const refreshModels = useCallback(() => {
    setModelsLoading(true)
    getModels(token)
      .then((data) => setModels(Array.isArray(data) ? data : []))
      .catch(() => {})
      .finally(() => setModelsLoading(false))
  }, [token])

  useEffect(() => {
    refreshNodes()
    refreshModels()
    fetchDatasetInfo().then(setDatasetInfo).catch(() => {})
    listDatasets(token).then(setDatasets).catch(() => {})
    getDatasetStatus(token).then(setMongoStatus).catch(() => {})
  }, [refreshNodes, refreshModels, token])

  useEffect(() => {
    if (!wsMessage) return
    const { type, jobId, status, progress, phase, node, mae, rmse, r2, message } = wsMessage

    if (jobRef.current && jobId && jobId !== jobRef.current) return

    if (type === 'jobUpdate') {
      setJob((prev) => prev ? { ...prev, status, progress: progress ?? prev.progress } : prev)
      if (status === 'done') { setTraining(false); refreshModels() }
      if (status === 'failed') setTraining(false)
    }
    if (type === 'phase') {
      setJob((prev) => prev ? { ...prev, phase } : prev)
    }
    if (type === 'nodeStart') {
      setJob((prev) => prev ? { ...prev, activeNodes: [...(prev.activeNodes || []), node] } : prev)
    }
    if (type === 'nodeDone') {
      setJob((prev) => prev ? { ...prev, doneNodes: [...(prev.doneNodes || []), node], progress: progress ?? prev.progress } : prev)
    }
    if (type === 'trainingComplete') {
      setJob((prev) => prev ? { ...prev, status: 'done', progress: 100, mae, rmse, r2 } : prev)
      setTraining(false)
      refreshModels()
    }
    if (type === 'error') {
      setJob((prev) => prev ? { ...prev, status: 'failed', errorMsg: message } : prev)
      setTraining(false)
    }
    if (type === 'migrateStart') {
      setMigrating(true)
      setMigrateProgress({ inserted: 0, progress: 0, done: false, error: null })
    }
    if (type === 'migrateProgress') {
      setMigrateProgress((prev) => ({ ...prev, inserted: wsMessage.inserted, progress: wsMessage.progress }))
    }
    if (type === 'migrateDone') {
      setMigrating(false)
      setMigrateProgress({ inserted: wsMessage.inserted, progress: 100, done: true, error: null })
      getDatasetStatus(token).then(setMongoStatus).catch(() => {})
    }
    if (type === 'migrateError') {
      setMigrating(false)
      setMigrateProgress((prev) => ({ ...prev, done: true, error: wsMessage.message }))
    }
  }, [wsMessage, refreshModels, token])

  async function handleAddNode() {
    const addr = newNodeAddr.trim()
    if (!addr) return
    if (!/^.+:\d+$/.test(addr)) {
      setNodeError('Formato inválido. Usa host:puerto (ej: 192.168.1.42:9000)')
      return
    }
    setNodeSaving(true)
    setNodeError('')
    try {
      const current = clusterData.nodes.map((n) => n.addr)
      if (current.includes(addr)) {
        setNodeError('Ese nodo ya está en la lista')
        return
      }
      const updated = [...current, addr]
      const data = await setClusterNodes(updated, token)
      setClusterData((prev) => ({ ...prev, ...data }))
      setNewNodeAddr('')
      refreshNodes()
      toast?.success(`Nodo ${addr} agregado`)
    } catch (err) {
      setNodeError(err.message)
    } finally {
      setNodeSaving(false)
    }
  }

  async function handleRemoveNode(addr) {
    setNodeSaving(true)
    setNodeError('')
    try {
      const updated = clusterData.nodes.map((n) => n.addr).filter((a) => a !== addr)
      if (updated.length === 0) {
        await resetClusterNodes(token)
        setClusterData((prev) => ({ ...prev, usingDynamic: false }))
      } else {
        const data = await setClusterNodes(updated, token)
        setClusterData((prev) => ({ ...prev, ...data }))
      }
      refreshNodes()
    } catch (err) {
      setNodeError(err.message)
    } finally {
      setNodeSaving(false)
    }
  }

  async function handleResetNodes() {
    setNodeSaving(true)
    setNodeError('')
    try {
      await resetClusterNodes(token)
      refreshNodes()
      toast?.info('Cluster restaurado a nodos Docker por defecto')
    } catch (err) {
      setNodeError(err.message)
    } finally {
      setNodeSaving(false)
    }
  }

  async function handleMigrate() {
    if (migrating) return
    setMigrateProgress(null)
    try {
      await migrateDataset(token)
    } catch (err) {
      setMigrateProgress({ inserted: 0, progress: 0, done: true, error: err.message })
      toast?.error(`Error al migrar: ${err.message}`)
    }
  }

  async function handleSelectDataset(path) {
    setDatasetSaving(true)
    try {
      await selectDataset(path, token)
      const [info, list] = await Promise.all([fetchDatasetInfo(), listDatasets(token)])
      setDatasetInfo(info)
      setDatasets(list)
      toast?.success('Dataset seleccionado')
    } catch (err) {
      toast?.error(`Error al seleccionar dataset: ${err.message}`)
    } finally {
      setDatasetSaving(false)
    }
  }

  async function handleTrain(e) {
    e.preventDefault()
    if (training) return
    setTraining(true)
    setJob(null)
    try {
      const { jobId } = await startTraining(
        { solver, lambda: parseFloat(lambda) || 1.0, maxRows: parseInt(maxRows, 10) || 0 },
        token,
      )
      jobRef.current = jobId
      setJob({ jobId, status: 'pending', progress: 0, phase: null, activeNodes: [], doneNodes: [] })
      toast?.info(`Entrenamiento iniciado con ${nodes.length} nodo${nodes.length !== 1 ? 's' : ''}`)
    } catch (err) {
      setTraining(false)
      setJob({ status: 'failed', errorMsg: err.message })
      toast?.error(`Error al iniciar entrenamiento: ${err.message}`)
    }
  }

  const phaseLabel = {
    loadingData: 'Cargando datos de MongoDB',
    distributing: 'Distribuyendo a nodos ML',
    solving: 'Resolviendo ecuaciones normales',
  }

  const nodes = clusterData.nodes || []
  const usingDynamic = clusterData.usingDynamic

  const step1Done = datasets.some((d) => d.isSelected)
  const step2Done = mongoStatus.ready
  const activeStep = !step1Done ? 1 : !step2Done ? 2 : 3

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal modal--wide" onClick={(e) => e.stopPropagation()}>
        <div className="modal__header">
          <div className="modal__header-left">
            <div className="modal__header-icon">⚙</div>
            <div>
              <h2>Panel de Administración</h2>
              <p>Gestión del cluster distribuido y entrenamiento de modelos ML</p>
            </div>
          </div>
          <button className="modal__close" onClick={onClose} aria-label="Cerrar">✕</button>
        </div>

        <div className="admin-panel">

          {/* ── Cluster de Nodos ── */}
          <section className="admin-section">
            <div className="admin-section__header">
              <div className="admin-section__title">
                <div className="admin-section__icon">🖥</div>
                <div>
                  <h3>
                    Cluster de Nodos ML
                    <span className={`cluster-mode ${usingDynamic ? 'cluster-mode--dynamic' : 'cluster-mode--default'}`}>
                      {usingDynamic ? 'Personalizado' : 'Docker por defecto'}
                    </span>
                  </h3>
                  <p>Nodos TCP que reciben particiones de datos para el entrenamiento distribuido</p>
                </div>
              </div>
              <button className="btn btn--sm btn--ghost" onClick={refreshNodes} disabled={nodesLoading}>
                {nodesLoading ? '…' : '↺ Actualizar'}
              </button>
            </div>

            <div className="node-list">
              {nodesAuthError && (
                <p className="admin-empty" style={{ color: 'var(--color-error)' }}>
                  ⚠ Sesión expirada — cierra sesión y vuelve a entrar.
                </p>
              )}
              {!nodesAuthError && nodes.length === 0 && !nodesLoading && (
                <p className="admin-empty">Sin nodos detectados — ¿están corriendo los procesos ML?</p>
              )}
              {nodes.map((n) => (
                <NodeBadge key={n.addr} node={n} onRemove={handleRemoveNode} />
              ))}
            </div>

            <div className="node-add">
              <input
                className="form-input node-add__input"
                type="text"
                placeholder="host:puerto  (ej: 192.168.1.42:9000)"
                value={newNodeAddr}
                onChange={(e) => setNewNodeAddr(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleAddNode()}
                disabled={nodeSaving}
              />
              <button className="btn btn--sm btn--primary" onClick={handleAddNode} disabled={nodeSaving || !newNodeAddr.trim()}>
                + Agregar nodo
              </button>
              {usingDynamic && (
                <button className="btn btn--sm btn--ghost" onClick={handleResetNodes} disabled={nodeSaving}>
                  Restaurar Docker
                </button>
              )}
            </div>
            {nodeError && <p className="node-add__error">⚠ {nodeError}</p>}
            <p className="node-add__hint">
              Para agregar un nodo externo: corre <code>docker compose -f docker-compose.node.yml up</code> en la otra PC y agrega su IP aquí.
            </p>
          </section>

          {/* ── Dataset + Migración ── */}
          <section className="admin-section">
            <div className="admin-section__header">
              <div className="admin-section__title">
                <div className="admin-section__icon">🗄</div>
                <div>
                  <h3>Dataset de Entrenamiento</h3>
                  <p>Selecciona el CSV, cárgalo a MongoDB y luego inicia el entrenamiento</p>
                </div>
              </div>
            </div>

            <div className="step-flow">
              <span className={`step-badge ${activeStep === 1 ? 'step-badge--active' : step1Done ? 'step-badge--done' : ''}`}>
                {step1Done ? '✓' : '①'} Seleccionar CSV
              </span>
              <span className="step-arrow">›</span>
              <span className={`step-badge ${activeStep === 2 ? 'step-badge--active' : step2Done ? 'step-badge--done' : ''}`}>
                {step2Done ? '✓' : '②'} Cargar a MongoDB
              </span>
              <span className="step-arrow">›</span>
              <span className={`step-badge ${activeStep === 3 ? 'step-badge--active' : ''}`}>
                ③ Entrenar modelo
              </span>
            </div>

            {/* Paso 1 */}
            <div>
              <p className="step-label">① Seleccionar CSV</p>
              {datasets.length === 0 && (
                <p className="admin-empty">No se encontraron archivos CSV en el servidor.</p>
              )}
              {datasets.length > 0 && (
                <div className="dataset-selector">
                  {datasets.map((d) => (
                    <button
                      key={d.path}
                      className={`dataset-option ${d.isSelected ? 'dataset-option--active' : ''}`}
                      onClick={() => !d.isSelected && handleSelectDataset(d.path)}
                      disabled={datasetSaving || migrating}
                      title={d.path}
                    >
                      <span className="dataset-option__name">{d.name}</span>
                      <span className="dataset-option__size">{d.sizeMB} MB</span>
                      {d.isSelected && <span className="dataset-option__check">✓ seleccionado</span>}
                    </button>
                  ))}
                </div>
              )}
            </div>

            {/* Paso 2 */}
            <div>
              <p className="step-label">② Cargar en MongoDB</p>
              <div className="migrate-row">
                <div className="mongo-status">
                  <span className={`mongo-dot ${mongoStatus.ready ? 'mongo-dot--ok' : 'mongo-dot--empty'}`} />
                  {mongoStatus.ready
                    ? <span>{mongoStatus.measurementsCount.toLocaleString()} documentos en MongoDB</span>
                    : <span className="mongo-empty">MongoDB vacío — migra primero</span>
                  }
                </div>
                <button
                  className="btn btn--sm btn--primary"
                  onClick={handleMigrate}
                  disabled={migrating || !step1Done}
                  title={!step1Done ? 'Selecciona un CSV primero' : ''}
                >
                  {migrating ? '⟳ Migrando…' : '↑ Migrar CSV → MongoDB'}
                </button>
              </div>

              {migrateProgress && (
                <div className="migrate-progress">
                  <ProgressBar value={migrateProgress.progress} />
                  {migrateProgress.done && !migrateProgress.error && (
                    <p className="migrate-done">✓ {migrateProgress.inserted.toLocaleString()} documentos insertados correctamente</p>
                  )}
                  {migrateProgress.error && (
                    <p className="migrate-error">✕ {migrateProgress.error}</p>
                  )}
                  {!migrateProgress.done && (
                    <p className="migrate-count">⟳ {migrateProgress.inserted.toLocaleString()} documentos insertados…</p>
                  )}
                </div>
              )}
            </div>

            {/* Info del dataset */}
            {datasetInfo && (
              <div>
                <p className="step-label">Información del CSV seleccionado</p>
                <div className="dataset-grid">
                  <div className="dataset-stat">
                    <span className="dataset-stat__label">Archivo</span>
                    <span className="dataset-stat__value dataset-stat__value--path" title={datasetInfo.path}>
                      {datasetInfo.path.split('/').pop()}
                    </span>
                  </div>
                  <div className="dataset-stat">
                    <span className="dataset-stat__label">Tamaño</span>
                    <span className="dataset-stat__value">{datasetInfo.fileSizeMB} MB</span>
                  </div>
                  <div className="dataset-stat">
                    <span className="dataset-stat__label">Filas</span>
                    <span className="dataset-stat__value">{datasetInfo.totalRows.toLocaleString()}</span>
                  </div>
                  <div className="dataset-stat">
                    <span className="dataset-stat__label">Columnas</span>
                    <span className="dataset-stat__value">{datasetInfo.columns}</span>
                  </div>
                  <div className="dataset-stat">
                    <span className="dataset-stat__label">Años</span>
                    <span className="dataset-stat__value">{datasetInfo.yearMin} – {datasetInfo.yearMax}</span>
                  </div>
                  <div className="dataset-stat">
                    <span className="dataset-stat__label">Media global</span>
                    <span className="dataset-stat__value">{datasetInfo.globalMean?.toFixed(4)}</span>
                  </div>
                </div>
                <div className="dataset-pollutants">
                  {(datasetInfo.pollutants || []).map((p) => (
                    <div key={p.name} className="dataset-pollutant-badge">
                      <span className="dataset-pollutant-badge__name">{p.name}</span>
                      <span className="dataset-pollutant-badge__rows">{p.rows.toLocaleString()} obs.</span>
                    </div>
                  ))}
                </div>
                <details className="dataset-columns">
                  <summary>Ver {datasetInfo.columns} columnas del dataset</summary>
                  <div className="dataset-columns__list">
                    {(datasetInfo.columnNames || []).map((c) => (
                      <code key={c} className="dataset-columns__col">{c}</code>
                    ))}
                  </div>
                </details>
              </div>
            )}
          </section>

          {/* ── Entrenamiento ── */}
          <section className="admin-section">
            <div className="admin-section__header">
              <div className="admin-section__title">
                <div className="admin-section__icon admin-section__icon--purple">🧠</div>
                <div>
                  <h3>③ Iniciar Entrenamiento Distribuido</h3>
                  <p>Regresión lineal distribuida entre {nodes.length} nodo{nodes.length !== 1 ? 's' : ''} via TCP + gob</p>
                </div>
              </div>
            </div>

            {!mongoStatus.ready && (
              <div className="app__banner app__banner--warning" style={{ position: 'static', transform: 'none', maxWidth: 'none' }}>
                ⚠ MongoDB vacío — completa los pasos anteriores antes de entrenar.
              </div>
            )}

            <form className="train-form" onSubmit={handleTrain}>
              <label className="form-label">
                Solver
                <select className="form-input" value={solver} onChange={(e) => setSolver(e.target.value)} disabled={training}>
                  <option value="ridge">Ridge (recomendado)</option>
                  <option value="svd">SVD</option>
                  <option value="normal">Ecuaciones normales</option>
                </select>
              </label>
              <label className="form-label">
                Lambda (λ)
                <input className="form-input" type="number" min="0" step="0.1" value={lambda} onChange={(e) => setLambda(e.target.value)} disabled={training} />
              </label>
              <label className="form-label">
                Máx. filas (0 = todas)
                <input className="form-input" type="number" min="0" step="10000" value={maxRows} onChange={(e) => setMaxRows(e.target.value)} disabled={training} />
              </label>
              <button
                className="btn btn--primary"
                type="submit"
                disabled={training || !mongoStatus.ready}
                title={!mongoStatus.ready ? 'Primero migra el dataset a MongoDB' : ''}
              >
                {training ? '⟳ Entrenando…' : `▶ Iniciar con ${nodes.length} nodo${nodes.length !== 1 ? 's' : ''}`}
              </button>
            </form>

            {job && (
              <div className="job-status">
                <div className="job-status__row">
                  <span className="job-status__label">Job ID</span>
                  <code title={job.jobId}>{job.jobId?.slice(0, 18)}…</code>
                </div>
                <div className="job-status__row">
                  <span className="job-status__label">Estado</span>
                  <span className={`job-badge job-badge--${job.status}`}>{job.status}</span>
                </div>
                {job.phase && (
                  <div className="job-status__row">
                    <span className="job-status__label">Fase actual</span>
                    <span>{phaseLabel[job.phase] || job.phase}</span>
                  </div>
                )}
                <ProgressBar value={job.progress || 0} />
                {job.doneNodes?.length > 0 && (
                  <div className="job-status__row">
                    <span className="job-status__label">Nodos completados</span>
                    <span>{job.doneNodes.join(', ')}</span>
                  </div>
                )}
                {job.status === 'done' && (
                  <div className="job-metrics">
                    <div className="metric"><span className="metric__label">R²</span><span className="metric__value">{job.r2?.toFixed(4)}</span></div>
                    <div className="metric"><span className="metric__label">MAE</span><span className="metric__value">{job.mae?.toFixed(4)}</span></div>
                    <div className="metric"><span className="metric__label">RMSE</span><span className="metric__value">{job.rmse?.toFixed(4)}</span></div>
                  </div>
                )}
                {job.errorMsg && <div className="job-status__error">✕ {job.errorMsg}</div>}
              </div>
            )}
          </section>

          {/* ── Modelos entrenados ── */}
          <section className="admin-section">
            <div className="admin-section__header">
              <div className="admin-section__title">
                <div className="admin-section__icon admin-section__icon--success">📈</div>
                <div>
                  <h3>Modelos Entrenados</h3>
                  <p>Historial de modelos generados por el cluster</p>
                </div>
              </div>
              <button className="btn btn--sm btn--ghost" onClick={refreshModels} disabled={modelsLoading}>
                {modelsLoading ? '…' : '↺ Actualizar'}
              </button>
            </div>
            {models.length === 0 && !modelsLoading && (
              <p className="admin-empty">No hay modelos aún. Inicia el primer entrenamiento.</p>
            )}
            {models.length > 0 && (
              <div className="table-scroll">
                <table className="models-table">
                  <thead>
                    <tr>
                      <th>Job ID</th><th>Solver</th><th>R²</th><th>MAE</th><th>RMSE</th><th>Filas train</th><th>Fecha</th>
                    </tr>
                  </thead>
                  <tbody>
                    {models.map((m) => <ModelRow key={m.jobId || m._id} model={m} />)}
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
