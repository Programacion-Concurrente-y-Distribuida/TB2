#!/usr/bin/env bash
# Levanta una instancia local de MongoDB en Docker con datos persistentes.
#
# Uso:
#   ./start_mongo.sh
#
# Variables de entorno opcionales:
#   MONGO_CONTAINER   nombre del contenedor   (default: aqs-mongo)
#   MONGO_PORT        puerto en el host       (default: 27017)
#   MONGO_VOLUME      volumen de datos        (default: aqs-mongo-data)
#   MONGO_VERSION     tag de la imagen        (default: 7)
#
# Tras ejecutarlo, la URL de conexion sera:
#   mongodb://localhost:27017
set -euo pipefail

MONGO_CONTAINER="${MONGO_CONTAINER:-aqs-mongo}"
MONGO_PORT="${MONGO_PORT:-27017}"
MONGO_VOLUME="${MONGO_VOLUME:-aqs-mongo-data}"
MONGO_VERSION="${MONGO_VERSION:-7}"

# Verifica que Docker este disponible y corriendo.
if ! docker info >/dev/null 2>&1; then
  echo "error: Docker no esta corriendo. Abre Docker Desktop e intenta de nuevo." >&2
  exit 1
fi

# Si el contenedor ya existe, lo (re)iniciamos en lugar de recrearlo:
# asi conservamos el volumen y la configuracion.
if docker ps -a --format '{{.Names}}' | grep -qx "${MONGO_CONTAINER}"; then
  if docker ps --format '{{.Names}}' | grep -qx "${MONGO_CONTAINER}"; then
    echo "MongoDB ya esta corriendo en el contenedor '${MONGO_CONTAINER}'."
  else
    echo "Iniciando contenedor existente '${MONGO_CONTAINER}'..."
    docker start "${MONGO_CONTAINER}" >/dev/null
  fi
else
  echo "Creando y levantando MongoDB ${MONGO_VERSION} (contenedor '${MONGO_CONTAINER}')..."
  docker run -d \
    --name "${MONGO_CONTAINER}" \
    --restart unless-stopped \
    -p "${MONGO_PORT}:27017" \
    -v "${MONGO_VOLUME}:/data/db" \
    "mongo:${MONGO_VERSION}" >/dev/null
fi

# Espera a que MongoDB acepte conexiones.
echo -n "Esperando a que MongoDB este listo"
for _ in $(seq 1 30); do
  if docker exec "${MONGO_CONTAINER}" mongosh --quiet --eval 'db.runCommand({ ping: 1 }).ok' >/dev/null 2>&1; then
    echo " listo."
    echo
    echo "MongoDB activo:"
    echo "  URL:        mongodb://localhost:${MONGO_PORT}"
    echo "  Contenedor: ${MONGO_CONTAINER}"
    echo "  Volumen:    ${MONGO_VOLUME} (datos persistentes)"
    echo
    echo "Para migrar el dataset:"
    echo "  go run ./cmd/migrate-mongo -mongo-url \"mongodb://localhost:${MONGO_PORT}\""
    exit 0
  fi
  echo -n "."
  sleep 1
done

echo
echo "error: MongoDB no respondio a tiempo. Revisa: docker logs ${MONGO_CONTAINER}" >&2
exit 1
