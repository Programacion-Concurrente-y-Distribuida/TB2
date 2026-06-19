#!/usr/bin/env bash
# Detiene la instancia local de MongoDB en Docker SIN borrar los datos.
#
# Uso:
#   ./stop_mongo.sh           # detiene el contenedor (datos intactos)
#   ./stop_mongo.sh --purge   # detiene Y elimina contenedor + volumen (borra todo)
set -euo pipefail

MONGO_CONTAINER="${MONGO_CONTAINER:-aqs-mongo}"
MONGO_VOLUME="${MONGO_VOLUME:-aqs-mongo-data}"

if docker ps --format '{{.Names}}' | grep -qx "${MONGO_CONTAINER}"; then
  echo "Deteniendo '${MONGO_CONTAINER}'..."
  docker stop "${MONGO_CONTAINER}" >/dev/null
  echo "Detenido. Los datos siguen en el volumen '${MONGO_VOLUME}'."
else
  echo "El contenedor '${MONGO_CONTAINER}' no esta corriendo."
fi

if [[ "${1:-}" == "--purge" ]]; then
  echo "Eliminando contenedor y volumen (esto BORRA todos los datos)..."
  docker rm "${MONGO_CONTAINER}" >/dev/null 2>&1 || true
  docker volume rm "${MONGO_VOLUME}" >/dev/null 2>&1 || true
  echo "Contenedor y volumen eliminados."
fi
