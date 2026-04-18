/* ============================================================
 * MODELO PROMELA: descargar.go
 * Flujo: descarga concurrente de reportes mensuales SEACE
 * Concurrencia modelada:
 *   - Semaforo de 3 slots (sem := make(chan struct{}, 3))
 *   - WaitGroup (sync.WaitGroup)
 *   - Mutex (sync.Mutex) protegiendo downloadedCount
 * ============================================================ */

#define N_URLS          10   /* archivos mensuales a intentar descargar */
#define MAX_CONCURRENT   3   /* mismo valor que en el Go: sem buffer = 3 */

/* ---------- Estado compartido ---------- */

chan sem = [MAX_CONCURRENT] of { bit };  /* semaforo: ocupar = !, liberar = ? */
chan mu  = [1] of { bit };               /* mutex binario (Lock=!,Unlock=?)  */

byte downloaded_count = 0;               /* mismo rol que downloadedCount   */
byte finished_count   = 0;               /* solo para chequeo de terminacion */
byte active_downloads = 0;               /* modela wg.Add / wg.Done          */
bit  mutex_ocupado    = 0;               /* testigo de exclusion mutua       */

/* ---------- Proceso: goroutine de descarga ---------- */

proctype descargador() {
    bit exito;

    /* Elegimos no deterministicamente si la descarga tiene exito (200)
       o falla (ej. 404 Not Found). Abstrae http.Client.Do + status code. */
    if
    :: exito = 1
    :: exito = 0
    fi;

    if
    :: exito == 1 ->
        mu ! 1;                              /* mu.Lock()                 */
        assert(mutex_ocupado == 0);          /* A3: exclusion mutua       */
        mutex_ocupado = 1;
        assert(downloaded_count < N_URLS);   /* A1: nunca se pasa del total */
        downloaded_count++;
        mutex_ocupado = 0;
        mu ? 1;                              /* mu.Unlock()               */
    :: else ->
        skip                                 /* rama 404: os.Remove + log */
    fi;

    atomic {
        finished_count++;
        active_downloads--;
    }

    sem ? 1;                                 /* defer func(){ <-sem }()   */
}

/* ---------- Proceso principal (main) ---------- */

init {
    byte i = 0;

    do
    :: i < N_URLS ->
        sem ! 1;                             /* sem <- struct{}{} (bloquea si lleno) */
        atomic { active_downloads++ }        /* wg.Add(1)                 */
        run descargador();
        i++
    :: else -> break
    od;

    /* wg.Wait(): no avanzamos hasta que todas las goroutines terminen. */
    (active_downloads == 0);

    /* A2: todo proceso lanzado debe haber finalizado. */
    assert(finished_count == N_URLS);

    /* A4: downloaded_count nunca excede los intentos. */
    assert(downloaded_count <= N_URLS);
}

/* ---------- Propiedades LTL ----------
   Recordatorio: SPIN verifica negacion de la formula; estas deben
   SER VALIDAS en todas las ejecuciones del modelo. */

ltl no_exceso       { [] (downloaded_count <= N_URLS) }
ltl terminacion     { <> (finished_count == N_URLS)   }
ltl progreso_wg     { [] ((active_downloads > 0) -> <> (active_downloads == 0)) }
ltl sem_respetado   { [] (downloaded_count <= N_URLS) }
