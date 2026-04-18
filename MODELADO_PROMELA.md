# Modelado Promela / SPIN: unión y limpieza de datos SEACE

Este documento describe el **modelado formal en Promela** del programa concurrente `1. Limpieza/union_y_limpieza.go`, verificable con **SPIN** mediante `1. Limpieza/union_y_limpieza.pml`. El resto del pipeline (descarga, expansión secuencial, etc.) queda fuera de alcance aquí.

---

## Parámetros del modelo

Constantes acotadas para que SPIN explore el espacio de estados en tiempo razonable (el modelo real en Go usa `runtime.NumCPU()` y tantos archivos como devuelva `filepath.Glob`):

| Constante | Valor en `.pml` | Rol |
|-----------|-----------------|-----|
| `N_WORKERS` | 3 | Pool de workers (análogo a `runtime.NumCPU()`). |
| `N_ARCHIVOS` | 6 | Cantidad de CSV en `data/` a procesar. |
| `MAX_FILAS` | 2 | Tope simbólico de filas válidas por archivo tras `procesarArchivo` (abstrae parseo, filtros, etc.). |

---

## Qué fragmentos del Go se modelan

La concurrencia relevante está en `main`: arranque de workers, envío de trabajos por canal, cierre de `jobs`, espera con `WaitGroup`, cierre de `results` y fusión bajo mutex.

**Pool de workers y envío de trabajos:**

```248:258:1. Limpieza/union_y_limpieza.go
	// 1. Iniciar Workers
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg)
	}

	// 2. Enviar Trabajos
	for _, file := range files {
		jobs <- Job{Path: file}
	}
	close(jobs)
```

**Goroutine que equivale a `wg.Wait()` seguido de `close(results)`:**

```265:269:1. Limpieza/union_y_limpieza.go
	// Goroutine para esperar y cerrar resultados
	go func() {
		wg.Wait()
		close(results)
	}()
```

**Recolector: mutex sobre `finalDataset` y `totalRecords`:**

```271:279:1. Limpieza/union_y_limpieza.go
	for res := range results {
		if res.Error != nil {
			fmt.Printf(" [ERR] %s: %v\n", res.File, res.Error)
			continue
		}
		mu.Lock()
		finalDataset = append(finalDataset, res.Rows...)
		totalRecords += len(res.Rows)
		mu.Unlock()
```

El cuerpo de `worker` solo encadena `procesarArchivo` y envío al canal `results`; la lógica secuencial de limpieza no introduce estado compartido adicional:

```112:118:1. Limpieza/union_y_limpieza.go
func worker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		rows, err := procesarArchivo(job.Path)
		results <- Result{Rows: rows, Error: err, File: filepath.Base(job.Path)}
	}
}
```

En el `.pml`, `procesarArchivo` se resume en una elección no determinista del número de filas emitidas (`0` … `MAX_FILAS`). La escritura a disco del CSV final no se modela (sin concurrencia relevante).

---

## Assertions en `union_y_limpieza.pml`

| # | Assertion | Qué detectaría si fallara |
|---|-----------|---------------------------|
| 1 | `mutex_ocupado == 0` al entrar en sección crítica del recolector | Fallo del esquema de exclusión mutua. |
| 2 | `jobs_enviados == N_ARCHIVOS` al final del recolector | No se enviaron todos los trabajos esperados. |
| 3 | `jobs_procesados == N_ARCHIVOS` | Algún job no fue consumido por un worker. |
| 4 | `total_records == finalDataset_size` | Desincronización entre contador y acumulado (carrera en el modelo). |

---

## Propiedades LTL

| Nombre | Idea |
|--------|------|
| `todos_procesados` | Eventualmente `jobs_procesados == N_ARCHIVOS`. |
| `cada_job_se_procesa` | Si ya se enviaron todos los jobs, eventualmente todos se procesan. |
| `progreso_workers` | Si hay workers vivos, eventualmente `workers_vivos == 0`. |
| `invariante_dataset` | En todo estado: `total_records == finalDataset_size`. |

Justificación en el dominio SEACE: terminación del pool, drenaje del canal tras `close(jobs)`, y coherencia del conteo de filas consolidadas respecto al dataset en memoria antes de escribir `dataset_limpio.csv`.

---

## Correspondencia Go ↔ Promela (resumen)

| Go | Promela (idea) |
|----|----------------|
| `jobs` / `results` bufferizados | Canales `jobs` y `results` de capacidad `N_ARCHIVOS`; valores abstraídos a `byte`. |
| `close(jobs)` | Bandera `jobs_closed` + condición `empty(jobs)`. |
| `wg.Wait(); close(results)` | `cerrador`: espera `workers_vivos == 0`, luego `results_closed`. |
| `for res := range results` | `recolector` consume hasta `empty(results) && results_closed`. |
| `mu.Lock` / `Unlock` al fusionar filas | `mu ! 1` / `mu ? 1` y variables `total_records`, `finalDataset_size`. |
| `procesarArchivo` | `if :: filas = 0 … :: filas = MAX_FILAS fi` |

---

## Verificación con SPIN

Compilar y ejecutar desde `1. Limpieza/`:

```bash
spin -a union_y_limpieza.pml && gcc -o pan pan.c
```

SPIN incorpora **varias** fórmulas LTL; en cada corrida solo se verifica **una**. Ejecutar una vez por propiedad:

```bash
./pan -a -N todos_procesados
./pan -a -N cada_job_se_procesa
./pan -a -N progreso_workers
./pan -a -N invariante_dataset
```

**Nota:** Con `N_ARCHIVOS = 6`, `N_WORKERS = 3` y `MAX_FILAS = 2`, el espacio de estados es más acotado que con valores mayores; aun así, la no determinista en `filas` puede hacer crecer la exploración. Si hace falta más rapidez, reducir otra vez esas constantes en el `.pml`.

### Resultado esperado

- `errors: 0` para cada `-N` elegido.
- Sin `assertion violated` ni trail de error para esas corridas.

### Por qué el modelo debería cumplir las propiedades

1. El canal `jobs` tiene capacidad `N_ARCHIVOS`, de modo que el envío secuencial no bloquea antes de `jobs_closed`.
2. Los workers solo salen cuando `empty(jobs) && jobs_closed`, lo que exige consumir todos los jobs.
3. El `cerrador` solo cierra la lógica de `results` cuando no quedan workers vivos, alineado con `wg.Wait()` antes de `close(results)`.
4. `total_records` y `finalDataset_size` se actualizan en la misma sección crítica en el recolector.

Si fallara la verificación, causas típicas serían: workers que abandonan el bucle sin procesar todos los jobs, `close(results)` demasiado pronto, o actualización de contadores fuera del mutex (en el modelo Promela actual van emparejados).
