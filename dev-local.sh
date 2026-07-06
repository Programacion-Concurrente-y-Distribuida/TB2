#!/usr/bin/env bash
# Levanta el sistema distribuido localmente:
#   - Mongo + Redis en Docker
#   - 3 nodos ML en puertos 9001, 9002, 9003
#   - API coordinador en :8080
# Uso: ./dev-local.sh
# Para detener todo: Ctrl+C

set -e
REPO="$(cd "$(dirname "$0")" && pwd)"
ML_DIR="$REPO/ML_and_concurrent_model"

# ── Infraestructura Docker ───────────────────────────────────────────────────
echo "▶ Levantando Mongo y Redis en Docker..."
docker compose -f "$REPO/docker-compose.yml" up -d mongo redis
echo "  Esperando que estén listos..."
sleep 4

# ── Compilar binarios ────────────────────────────────────────────────────────
echo "▶ Compilando..."
cd "$ML_DIR"
go build -o /tmp/ml-node   ./cmd/ml-node/
go build -o /tmp/ml-api    ./cmd/api/
echo "  Listo."

# ── Nodos ML ─────────────────────────────────────────────────────────────────
echo "▶ Iniciando nodos ML en puertos 9001, 9002, 9003..."
NODE_PORT=9001 /tmp/ml-node &  NODE1_PID=$!
NODE_PORT=9002 /tmp/ml-node &  NODE2_PID=$!
NODE_PORT=9003 /tmp/ml-node &  NODE3_PID=$!
sleep 1

# ── API coordinador ──────────────────────────────────────────────────────────
echo "▶ Iniciando API en :8080..."
export PORT=":8080"
export MONGO_URL="mongodb://localhost:27017"
export MONGO_DB="aqs"
export REDIS_ADDR="localhost:6379"
export JWT_SECRET="${JWT_SECRET:-dev_secret_local}"
export ADMIN_USER="${ADMIN_USER:-admin}"
export ADMIN_PASS="${ADMIN_PASS:-admin123}"
export NODE_ADDRS="localhost:9001,localhost:9002,localhost:9003"
export INPUT_PATH="$REPO/data/dataset.csv"

/tmp/ml-api &  API_PID=$!

# ── Frontend ──────────────────────────────────────────────────────────────────
echo "▶ Iniciando frontend en :5173..."
cd "$REPO/frontend"
npm install --silent 2>/dev/null || true
npm run dev -- --port 5173 &  FRONT_PID=$!
cd "$REPO"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Nodo 1 → localhost:9001  (PID $NODE1_PID)"
echo "  Nodo 2 → localhost:9002  (PID $NODE2_PID)"
echo "  Nodo 3 → localhost:9003  (PID $NODE3_PID)"
echo "  API    → http://localhost:8080  (PID $API_PID)"
echo "  UI     → http://localhost:5173  (PID $FRONT_PID)"
echo "  Login  → admin / admin123"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Ctrl+C para detener todo"
echo ""

# ── Cleanup al salir ─────────────────────────────────────────────────────────
trap 'echo ""; echo "Deteniendo procesos..."; kill $NODE1_PID $NODE2_PID $NODE3_PID $API_PID $FRONT_PID 2>/dev/null; docker compose -f "$REPO/docker-compose.yml" stop mongo redis; echo "Listo."' EXIT INT TERM

wait $API_PID
