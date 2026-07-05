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
      setError(err.message || 'Credenciales incorrectas')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="login-card" onClick={(e) => e.stopPropagation()}>
        <div className="login-card__top">
          <div className="login-card__icon">🔐</div>
          <h2>Acceso de Administrador</h2>
          <p>Gestión del cluster y entrenamiento de modelos</p>
        </div>

        <form className="login-card__body" onSubmit={handleSubmit}>
          {error && (
            <div className="modal__error">
              <span>⚠</span> {error}
            </div>
          )}

          <label className="form-label">
            Usuario
            <div className="input-with-icon">
              <span className="input-with-icon__icon">👤</span>
              <input
                className="form-input"
                type="text"
                autoComplete="username"
                placeholder="Nombre de usuario"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                disabled={loading}
              />
            </div>
          </label>

          <label className="form-label">
            Contraseña
            <div className="input-with-icon">
              <span className="input-with-icon__icon">🔑</span>
              <input
                className="form-input"
                type="password"
                autoComplete="current-password"
                placeholder="••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={loading}
              />
            </div>
          </label>

          <button className="login-card__submit" type="submit" disabled={loading}>
            {loading ? 'Verificando…' : 'Ingresar al sistema'}
          </button>
        </form>
      </div>
    </div>
  )
}
