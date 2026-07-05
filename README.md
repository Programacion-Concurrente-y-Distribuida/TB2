# Sistema Distribuido de Predicción de Calidad del Aire — Lima

Predice niveles de contaminación del aire (PM2.5, O3, NO2, CO) por distrito de Lima Metropolitana usando un modelo de regresión lineal distribuido en Go.

## Arquitectura

```
Frontend (React + Vite)
        │ HTTP/WebSocket
        ▼
  API Coordinador (Go :8080)
   ├── MongoDB  (datos históricos + modelos)
   ├── Redis    (estado de jobs + nodos dinámicos)
   └── Cluster ML (TCP)
        ├── Nodo 1 :9001  ─┐
        ├── Nodo 2 :9002   ├─ Regresión distribuida (XTX/XTy parciales via gob)
        └── Nodo 3 :9003  ─┘
```

El coordinador divide la matriz de diseño en particiones iguales, cada nodo computa su `XTX` y `XTy` parcial, el coordinador agrega y resuelve las ecuaciones normales (Ridge/SVD).

---

## Requisitos

| Herramienta | Versión mínima | Instalación |
|---|---|---|
| Go | 1.21 | `brew install go` |
| Node.js | 18 | `brew install node` |
| Docker | cualquiera | `brew install colima && colima start` |

---

## Inicio rápido (desarrollo local)

### 1. Clonar el repositorio

```bash
git clone <url-del-repo>
cd TB2-1
```

### 2. Dataset

El dataset (`aqs_final_3M.csv`, ~1.4 GB) **no está incluido en el repo** por su tamaño. Colócalo manualmente en:

```
TB2-1/data/aqs_final_3M.csv
```

> Fuentes: [EPA AQS](https://www.epa.gov/aqs) · [SENAMHI REMCA](https://www.datosabiertos.gob.pe/dataset/monitoreo-de-los-contaminantes-del-aire-en-lima-metropolitana-servicio-nacional-de)

### 3. Iniciar Docker (Colima en Mac)

```bash
colima start
export DOCKER_HOST="unix://${HOME}/.colima/docker.sock"
```

### 4. Levantar Mongo y Redis

```bash
docker run -d --name local-mongo -p 27017:27017 mongo:7
docker run -d --name local-redis -p 6379:6379 redis:7-alpine
```

> Si ya los tienes corriendo de una sesión anterior: `docker start local-mongo local-redis`

### 5. Migrar el dataset a MongoDB

```bash
cd ML_and_concurrent_model

go run ./cmd/migrate-mongo \
  -mongo-url "mongodb://localhost:27017" \
  -input "../data/aqs_final_3M.csv" \
  -db aqs \
  -collection measurements \
  -drop

cd ..
```

Este paso usa un pipeline concurrente productor/consumidor con `InsertMany` en paralelo. Con 3M de filas tarda ~3-5 minutos.

### 6. Levantar el backend (nodos ML + API)

```bash
./dev-local.sh
```

El script compila los binarios, levanta 3 nodos ML en puertos **9001, 9002, 9003** y la API en **:8080**. `Ctrl+C` detiene todo limpiamente.

### 7. Levantar el frontend

En otra terminal:

```bash
cd frontend
npm install
npm run dev
# → http://localhost:5173
```

### 8. Entrenar el modelo

1. Abre **http://localhost:5173**
2. Clic en **Admin** (esquina superior derecha)
3. Login: `admin` / `admin123`
4. En el panel de administración → **Iniciar Entrenamiento Distribuido**
5. Selecciona el solver (`Ridge` recomendado) y lanza
6. El progreso se muestra en tiempo real vía WebSocket
7. Al terminar, el mapa se llena automáticamente con predicciones por distrito

---

## Levantar en producción (Docker Compose completo)

```bash
export DOCKER_HOST="unix://${HOME}/.colima/docker.sock"  # solo Mac con Colima

# Ajusta las variables de entorno si es necesario
docker compose up --build
```

Servicios expuestos:
- Frontend: http://localhost:3000
- API: http://localhost:8080

Credenciales por defecto: `admin` / `admin123` (configurable via `ADMIN_USER` / `ADMIN_PASS`).

---

## Variables de entorno (API)

| Variable | Default | Descripción |
|---|---|---|
| `PORT` | `:8080` | Puerto de escucha |
| `MONGO_URL` | `mongodb://mongo:27017` | URI de MongoDB |
| `MONGO_DB` | `aqs` | Nombre de la base de datos |
| `REDIS_ADDR` | `redis:6379` | Dirección de Redis |
| `JWT_SECRET` | — | Secreto HMAC para tokens JWT (requerido) |
| `ADMIN_USER` | `admin` | Usuario administrador |
| `ADMIN_PASS` | — | Contraseña administrador (requerida) |
| `NODE_ADDRS` | `ml-node-1:9000` | Lista de nodos ML separada por comas |
| `INPUT_PATH` | `aqs_final_3M.csv` | Ruta del CSV para entrenamiento |

---

## Estructura del proyecto

```
TB2-1/
├── ML_and_concurrent_model/
│   ├── cmd/
│   │   ├── api/            # Servidor HTTP + coordinador
│   │   ├── ml-node/        # Nodo ML (servidor TCP)
│   │   ├── migrate-mongo/  # Migración CSV → MongoDB
│   │   ├── train-linear/   # Entrenamiento standalone
│   │   └── predict-linear/ # Predicción standalone
│   └── internal/
│       ├── apiserver/      # Handlers HTTP, WebSocket, auth
│       ├── aqsml/          # Pipeline ML (encoding, modelo, datos)
│       ├── coordinator/    # Orquestación distribuida TCP
│       └── protocol/       # Tipos gob (TrainTask / TrainResult)
├── frontend/               # SPA React (Vite + Leaflet)
├── docker/                 # Dockerfiles para API y nodos
├── data/                   # Dataset CSV (no incluido en git)
├── docker-compose.yml      # Stack completo
├── docker-compose.node.yml # Nodo ML adicional en red
└── dev-local.sh            # Script de desarrollo local
```

---

## Cómo funciona la distribución

1. El coordinador lee el CSV y construye la matriz de diseño (`X`, `y`)
2. Divide `X` en N particiones iguales (una por nodo)
3. Envía cada partición vía TCP usando serialización `encoding/gob`
4. Cada nodo computa su `XTX` parcial y `XTy` parcial de forma concurrente
5. El coordinador suma todos los parciales y resuelve `β = (XTX)⁻¹ XTy`
6. El modelo resultante se persiste en MongoDB

## Agregar un nodo ML adicional desde otra PC

```bash
# En la otra PC, levantar un nodo:
docker compose -f docker-compose.node.yml up

# En el panel Admin de la UI, agregar su IP:
# host:9000  (ej: 192.168.1.42:9000)
```
