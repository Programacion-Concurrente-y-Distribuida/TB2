# Frontend — Calidad del Aire, Lima

Mapa de predicción de contaminación del aire por distrito de Lima Metropolitana,
inspirado (de forma mucho más simple) en el
[mapa de pronóstico de calidad del aire de SENAMHI](https://www.senamhi.gob.pe/?p=calidad-del-aire-pronostico).

React + Vite + Leaflet. Consume el API Go de `ML_and_concurrent_model/`.

## Cómo correrlo

```bash
npm install
npm run dev
```

Por defecto `npm run dev` levanta Vite en `http://localhost:5173` y proxea
`/api` y `/ws` hacia `http://localhost:8080` (el API Go corriendo localmente
o vía `docker-compose up`). Si el API corre en otra URL, define
`VITE_API_PROXY_TARGET` antes de `npm run dev`, o copia `.env.example` a
`.env.local` y define `VITE_API_BASE_URL` para builds de producción.

```bash
npm run build   # genera dist/
```

## Arquitectura

- `src/components/MapView.jsx` — mapa Leaflet con los 43 distritos como capa
  GeoJSON (choropleth), coloreados según la predicción del contaminante
  seleccionado.
- `src/components/PollutantSelector.jsx` / `Legend.jsx` / `DistrictPanel.jsx`
  — sidebar: selector de contaminante, leyenda de niveles, panel de detalle
  del distrito seleccionado.
- `src/api/client.js` — cliente fetch contra el API.
- `src/data/lima-districts.geojson` — polígonos de los 43 distritos de la
  provincia de Lima.

El backend expone dos endpoints **públicos** (sin JWT, a diferencia de
`/api/predict` que es el endpoint interno de entrenamiento/panel admin):

- `GET /api/districts` — lista estática de distritos (id, nombre, centroide).
- `GET /api/districts/predict?pollutant=PM2.5` — corre el modelo entrenado más
  reciente sobre cada distrito y devuelve `{ id, name, lat, lon, prediction,
  unit, level, color }[]`. `pollutant` acepta `PM2.5` (default), `O3`, `NO2`,
  `CO`.

Ver `ML_and_concurrent_model/internal/apiserver/districts.go`.

## Datos geográficos

Límites distritales: **Instituto Geográfico Nacional (IGN)** del Perú, vía
OCHA / Humanitarian Data Exchange (dataset "Peru - Subnational Administrative
Boundaries", COD-AB). Licencia **CC BY-IGO** — requiere atribución (incluida
en el footer de la app). Los 43 distritos corresponden a la provincia de Lima
(no incluye Callao, que es una provincia constitucional aparte).

Los centroides usados como feature Latitude/Longitude para el modelo (en
`ML_and_concurrent_model/internal/apiserver/data/lima_districts.json`) se
calcularon a partir de esos mismos polígonos, para que mapa y predicciones
usen una única fuente de coordenadas.

## Limitación importante del modelo

El modelo de regresión lineal se entrena con el dataset EPA AQS de EE.UU.
(`aqs_clean.csv` / `aqs_final_3M.csv`), que no tiene estaciones en Perú. Para
poder pedirle una predicción por distrito de Lima, el endpoint
`/api/districts/predict` arma un input sintético por distrito:

- `Latitude`/`Longitude`/`Year` — reales, del distrito y el año actual.
- El resto de features numéricas (Observation Count, Valid Day Count, etc.)
  usan el **promedio histórico de ese contaminante** en el dataset de
  entrenamiento (no hay equivalente real para Lima).
- `State Code` se omite a propósito (Lima no tiene código de estado de
  EE.UU.), lo que el encoding trata como "sin categoría" — no aporta señal.

Esto es una simplificación metodológica razonable para una demo académica,
pero significa que la predicción refleja sobre todo el efecto de
latitud/longitud/contaminante aprendido por el modelo, no mediciones reales
de Lima. Si más adelante se consigue un dataset con estaciones peruanas
(ej. SENAMHI/MINAM), sería el reemplazo natural de `aqs_clean.csv` para que
las predicciones sean representativas de verdad.
