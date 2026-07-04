import { useEffect, useMemo, useState } from 'react'
import { MapContainer, TileLayer, GeoJSON } from 'react-leaflet'
import geojsonUrl from '../data/lima-districts.geojson?url'

const LIMA_CENTER = [-12.06, -77.03]

export default function MapView({ predictions, selectedId, onSelectDistrict }) {
  const [geojson, setGeojson] = useState(null)

  useEffect(() => {
    fetch(geojsonUrl)
      .then((res) => res.json())
      .then(setGeojson)
      .catch((err) => console.error('No se pudo cargar el GeoJSON de distritos', err))
  }, [])

  const predictionById = useMemo(() => {
    const map = new Map()
    for (const d of predictions) map.set(String(d.id), d)
    return map
  }, [predictions])

  function styleFeature(feature) {
    const pred = predictionById.get(String(feature.properties.id))
    const isSelected = pred && pred.id === selectedId
    return {
      fillColor: pred ? pred.color : '#cccccc',
      fillOpacity: isSelected ? 0.9 : 0.65,
      color: isSelected ? '#1d1d1d' : '#555555',
      weight: isSelected ? 2.5 : 1,
    }
  }

  function onEachFeature(feature, layer) {
    const pred = predictionById.get(String(feature.properties.id))
    if (pred) {
      layer.bindTooltip(`${pred.name}: ${pred.prediction.toFixed(1)} ${pred.unit} — ${pred.level}`)
    }
    layer.on('click', () => {
      if (pred) onSelectDistrict(pred)
    })
  }

  return (
    <MapContainer center={LIMA_CENTER} zoom={11} className="map-container">
      <TileLayer
        attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
      />
      {geojson && (
        <GeoJSON
          key={predictions.map((p) => `${p.id}:${p.color}`).join(',')}
          data={geojson}
          style={styleFeature}
          onEachFeature={onEachFeature}
        />
      )}
    </MapContainer>
  )
}
