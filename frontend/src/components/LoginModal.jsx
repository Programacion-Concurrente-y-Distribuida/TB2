import { useState } from 'react'
import { login } from '../api/client.js'

export default function LoginModal({ onLogin, onClose }) {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e) {
    e.preventDefault()
    if (!username || !password) {
      setError('Completa todos los campos')
      return
    }
    setLoading(true)
    setError('')
    try {
      const { token } = await login(username, password)
      onLogin(token)
    } catch (err) {
      setError(err.message || 'Error de autenticación')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal__header">
          <h2>Acceso de Administrador</h2>
          <button className="modal__close" onClick={onClose} aria-label="Cerrar">✕</button>
        </div>

        <form className="modal__body" onSubmit={handleSubmit}>
          {error && <p className="modal__error">{error}</p>}

          <label className="form-label">
            Usuario
            <input
              className="form-input"
              type="text"
              autoComplete="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              disabled={loading}
            />
          </label>

          <label className="form-label">
            Contraseña
            <input
              className="form-input"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={loading}
            />
          </label>

          <button className="btn btn--primary" type="submit" disabled={loading}>
            {loading ? 'Ingresando…' : 'Ingresar'}
          </button>
        </form>
      </div>
    </div>
  )
}
