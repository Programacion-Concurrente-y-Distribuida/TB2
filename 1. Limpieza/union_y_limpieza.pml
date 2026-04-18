/* ============================================================
 * MODELO PROMELA: union_y_limpieza.go
 * Flujo: worker-pool que limpia y consolida CSV mensuales SEACE
 * Concurrencia modelada:
 *   - Canal jobs    (buffered, tamano = #archivos)
 *   - Canal results (buffered, tamano = #archivos)
 *   - N_WORKERS goroutines worker()
 *   - Goroutine cerrador:  wg.Wait(); close(results)
 *   - Goroutine recolector (main loop: for res := range results)
 *   - Mutex mu protegiendo finalDataset y totalRecords
 * ============================================================ */

#define N_WORKERS    3    /* mismo rol que runtime.NumCPU()           */
#define N_ARCHIVOS   6    /* archivos .csv en la carpeta data/        */
#define MAX_FILAS    2    /* tope simbolico de filas limpias/archivo  */

/* ---------- Estado compartido ---------- */

chan jobs    = [N_ARCHIVOS] of { byte };  /* Job{Path}    (solo id)     */
chan results = [N_ARCHIVOS] of { byte };  /* Result.Rows  (solo conteo) */
chan mu      = [1]          of { bit  };  /* sync.Mutex                 */

byte jobs_enviados      = 0;
byte jobs_procesados    = 0;
byte total_records      = 0;
byte finalDataset_size  = 0;   /* duplica total_records para invariante */
byte workers_vivos      = 0;
bit  jobs_closed        = 0;   /* modela close(jobs)                    */
bit  results_closed     = 0;   /* modela close(results) tras wg.Wait()  */
bit  mutex_ocupado      = 0;   /* testigo de exclusion mutua            */

/* ---------- Worker ---------- */

proctype worker() {
    byte archivo;
    byte filas;

    do
    :: jobs ? archivo ->
        /* procesarArchivo(): eleccion no determinista del numero de
           filas validas resultantes (abstrae parseo, RUC, fechas,
           filtro TOTAL > 0, etc.). */
        if
        :: filas = 0
        :: filas = 1
        :: filas = 2
        :: filas = MAX_FILAS
        fi;
        results ! filas;
        atomic { jobs_procesados++ }
    :: empty(jobs) && jobs_closed -> break
    od;

    atomic { workers_vivos-- }
}

/* ---------- Goroutine cerradora (wg.Wait + close(results)) ---------- */

proctype cerrador() {
    (workers_vivos == 0);    /* wg.Wait()       */
    results_closed = 1;      /* close(results)  */
}

/* ---------- Recolector (for res := range results) ---------- */

proctype recolector() {
    byte filas;

    do
    :: results ? filas ->
        mu ! 1;                              /* mu.Lock()              */
        assert(mutex_ocupado == 0);          /* exclusion mutua        */
        mutex_ocupado = 1;
        finalDataset_size = finalDataset_size + filas;
        total_records    = total_records    + filas;
        mutex_ocupado = 0;
        mu ? 1;                              /* mu.Unlock()            */
    :: empty(results) && results_closed -> break
    od;

    /* Invariantes finales */
    assert(jobs_enviados   == N_ARCHIVOS);
    assert(jobs_procesados == N_ARCHIVOS);
    assert(total_records   == finalDataset_size);
}

/* ---------- main ---------- */

init {
    byte w = 0;
    byte i = 0;

    /* 1. Arrancar workers */
    do
    :: w < N_WORKERS ->
        atomic { workers_vivos++ }
        run worker();
        w++
    :: else -> break
    od;

    /* 2. Enviar jobs */
    do
    :: i < N_ARCHIVOS ->
        jobs ! i;
        atomic { jobs_enviados++ }
        i++
    :: else -> break
    od;
    jobs_closed = 1;                         /* close(jobs)            */

    /* 3. Goroutines auxiliares */
    run cerrador();
    run recolector();
}

/* ---------- Propiedades LTL ---------- */

ltl todos_procesados {
    <> (jobs_procesados == N_ARCHIVOS)
}

ltl cada_job_se_procesa {
    [] ( (jobs_enviados == N_ARCHIVOS) -> <> (jobs_procesados == N_ARCHIVOS) )
}

ltl progreso_workers {
    [] ( (workers_vivos > 0) -> <> (workers_vivos == 0) )
}

ltl invariante_dataset {
    [] (total_records == finalDataset_size)
}
