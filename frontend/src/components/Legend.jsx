import { AQI_LEVELS } from '../utils/aqi.js'

export default function Legend() {
  return (
    <div className="legend">
      <h3>Nivel</h3>
      <ul>
        {AQI_LEVELS.map((l) => (
          <li key={l.level}>
            <span className="swatch" style={{ backgroundColor: l.color }} />
            {l.level}
          </li>
        ))}
      </ul>
    </div>
  )
}
