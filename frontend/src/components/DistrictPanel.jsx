export default function DistrictPanel({ district }) {
  if (!district) {
    return (
      <div className="district-panel district-panel--empty">
        <p>Selecciona un distrito en el mapa para ver el detalle de la prediccion.</p>
      </div>
    )
  }

  return (
    <div className="district-panel">
      <h3>{district.name}</h3>
      <p className="district-panel__value" style={{ color: district.color }}>
        {district.prediction.toFixed(1)} {district.unit}
      </p>
      <p className="district-panel__level">{district.level}</p>
    </div>
  )
}
