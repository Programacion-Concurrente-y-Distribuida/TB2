import { POLLUTANTS } from '../utils/aqi.js'

export default function PollutantSelector({ value, onChange }) {
  return (
    <div className="pollutant-selector">
      <label htmlFor="pollutant">Contaminante</label>
      <select id="pollutant" value={value} onChange={(e) => onChange(e.target.value)}>
        {POLLUTANTS.map((p) => (
          <option key={p.id} value={p.id}>
            {p.label}
          </option>
        ))}
      </select>
    </div>
  )
}
