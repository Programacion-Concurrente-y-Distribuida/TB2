export default function DistrictPanel({ district }) {
  if (!district) {
    return (
      <div className="district-panel district-panel--empty">
        <p>📍 Haz clic en un distrito del mapa para ver la predicción de calidad del aire.</p>
      </div>
    )
  }

  return (
    <div className="district-panel" style={{ borderLeft: `3px solid ${district.color}` }}>
      <h3>{district.name}</h3>
      <div className="district-panel__value" style={{ color: district.color }}>
        {district.prediction.toFixed(1)}
        <span style={{ fontSize: '1rem', fontWeight: 500, marginLeft: '0.3rem', color: 'var(--color-text-muted)' }}>{district.unit}</span>
      </div>
      <div className="district-panel__level">{district.level}</div>
      <div className="district-panel__badge" style={{ backgroundColor: district.color }}>
        {district.level}
      </div>
    </div>
  )
}
