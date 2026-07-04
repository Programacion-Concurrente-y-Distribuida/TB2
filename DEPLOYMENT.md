# Despliegue distribuido — 3 dispositivos

## Arquitectura

| Dispositivo | IP | Rol | Servicios |
|---|---|---|---|
| Mac | `192.168.1.95` | Coordinador | API + MongoDB + Redis |
| PC Windows | `192.168.1.42` | Nodo ML 1 | ml-node puerto 9000 |
| Laptop Windows | `192.168.1.38` | Nodo ML 2 | ml-node puerto 9000 |

## Orden de arranque

```
1. Primero: los dos Windows (nodos deben estar listos antes que el coordinador intente conectarse)
2. Luego:   el Mac (API)
```

---

## Dispositivo 1 — Mac `192.168.1.95` (Coordinador)

> Corre la API, MongoDB y Redis.

**1. Verifica que el `.env` está correcto:**
```bash
cat .env
```
Debe mostrar:
```
JWT_SECRET=aqs_secreto_2026
ADMIN_USER=admin
ADMIN_PASS=admin123
NODE_ADDRS=192.168.1.38:9000,192.168.1.42:9000
```

**2. Verifica que el CSV está en `./data/`:**
```bash
ls -lh data/dataset.csv
```

**3. Levanta el coordinador:**
```bash
docker compose -f docker-compose.coordinator.yml --env-file .env up --build -d
```

**4. Verifica que todo levantó:**
```bash
docker compose -f docker-compose.coordinator.yml ps
```

**5. Prueba que la API responde:**
```bash
curl http://localhost:8080/api/health
```

---

## Dispositivo 2 — PC Windows `192.168.1.42` (Nodo ML 1)

> Prerequisito: tener [Docker Desktop](https://www.docker.com/products/docker-desktop/) instalado y corriendo.

**1. Abre PowerShell y clona el repo:**
```powershell
git clone https://github.com/Programacion-Concurrente-y-Distribuida/TB2.git
cd TB2
```

**2. Levanta el nodo:**
```powershell
docker compose -f docker-compose.node.yml up --build -d
```

**3. Verifica que está corriendo:**
```powershell
docker compose -f docker-compose.node.yml ps
docker compose -f docker-compose.node.yml logs
```
Debe mostrar: `ml-node ready on :9000`

**4. Abre el puerto 9000 en el firewall** (ejecutar como Administrador):
```powershell
netsh advfirewall firewall add rule name="ML Node 9000" dir=in action=allow protocol=TCP localport=9000
```

---

## Dispositivo 3 — Laptop Windows `192.168.1.38` (Nodo ML 2)

> Prerequisito: tener [Docker Desktop](https://www.docker.com/products/docker-desktop/) instalado y corriendo.

**1. Abre PowerShell y clona el repo:**
```powershell
git clone https://github.com/Programacion-Concurrente-y-Distribuida/TB2.git
cd TB2
```

**2. Levanta el nodo:**
```powershell
docker compose -f docker-compose.node.yml up --build -d
```

**3. Verifica que está corriendo:**
```powershell
docker compose -f docker-compose.node.yml ps
docker compose -f docker-compose.node.yml logs
```
Debe mostrar: `ml-node ready on :9000`

**4. Abre el puerto 9000 en el firewall** (ejecutar como Administrador):
```powershell
netsh advfirewall firewall add rule name="ML Node 9000" dir=in action=allow protocol=TCP localport=9000
```

---

## Verificación final — desde el Mac

Una vez los 3 dispositivos estén corriendo:

**1. Obtén el token de autenticación:**
```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r .token)
```

**2. Verifica que la API ve los dos nodos:**
```bash
curl -s http://localhost:8080/api/cluster/nodes \
  -H "Authorization: Bearer $TOKEN" | jq .
```
Respuesta esperada:
```json
[
  { "addr": "192.168.1.38:9000", "status": "ok", "latency_ms": 0 },
  { "addr": "192.168.1.42:9000", "status": "ok", "latency_ms": 0 }
]
```

**3. Lanza un entrenamiento distribuido:**
```bash
curl -s -X POST http://localhost:8080/api/train \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"solver":"ridge","lambda":1.0,"max_rows":100000}' | jq .
```

**4. Consulta el estado del job:**
```bash
curl -s http://localhost:8080/api/train/<job_id> \
  -H "Authorization: Bearer $TOKEN" | jq .
```

---

## Troubleshooting

**Nodo no aparece como `ok`:**
- Confirma que Docker está corriendo en el Windows
- Verifica la regla de firewall del paso 4
- Prueba conectividad desde el Mac: `nc -zv 192.168.1.42 9000`

**La API no levanta:**
```bash
docker compose -f docker-compose.coordinator.yml logs api
```

**MongoDB o Redis no responden:**
```bash
docker compose -f docker-compose.coordinator.yml logs mongo
docker compose -f docker-compose.coordinator.yml logs redis
```

**Apagar todo:**
```bash
# Mac
docker compose -f docker-compose.coordinator.yml down

# Windows
docker compose -f docker-compose.node.yml down
```
