// Mantener en sincronia con las etiquetas/colores definidos en
// ML_and_concurrent_model/internal/apiserver/districts.go — el backend ya
// calcula "level" y "color" por distrito; esto solo alimenta la leyenda.
export const POLLUTANTS = [
  { id: 'PM2.5', label: 'PM2.5', unit: 'µg/m³' },
  { id: 'O3', label: 'Ozono (O3)', unit: 'ppm' },
  { id: 'NO2', label: 'Dióxido de nitrógeno (NO2)', unit: 'ppb' },
  { id: 'CO', label: 'Monóxido de carbono (CO)', unit: 'ppm' },
]

export const AQI_LEVELS = [
  { level: 'Bueno', color: '#00e400' },
  { level: 'Moderado', color: '#ffff00' },
  { level: 'Dañino (grupos sensibles)', color: '#ff7e00' },
  { level: 'Dañino', color: '#ff0000' },
  { level: 'Muy dañino', color: '#8f3f97' },
  { level: 'Peligroso', color: '#7e0023' },
]
