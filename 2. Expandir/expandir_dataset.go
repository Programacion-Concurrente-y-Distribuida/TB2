package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// ─────────────────────────────────────────────
// CONFIGURACIÓN
// ─────────────────────────────────────────────

func main() {
	start := time.Now()
	rand.Seed(time.Now().UnixNano())

	// Rutas de archivos
	dataDir := "data"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		dataDir = "../data"
	}
	
	// Rutas de archivos: El dataset limpio esta en la raíz del proyecto (PC1)
	entrada := "dataset_limpio.csv"
	// Si no existe en la raíz, probamos un nivel arriba (por si se ejecuta desde la carpeta 2. Expandir)
	if _, err := os.Stat(entrada); os.IsNotExist(err) {
		entrada = "../dataset_limpio.csv"
	}
	
	salida := filepath.Join(filepath.Dir(entrada), "dataset_1m.csv")
	objetivo := 1000000

	fmt.Println("============================================================")
	fmt.Println(" [PCD] Iniciando Expansion Concurrente a 1 Millon")
	fmt.Println("============================================================")

	// 1. Leer dataset original completo a memoria
	f, err := os.Open(entrada)
	if err != nil {
		fmt.Printf("Error abriendo entrada: %v\n", err)
		return
	}
	reader := csv.NewReader(f)
	reader.Comma = ';'
	records, err := reader.ReadAll()
	f.Close()
	if err != nil {
		fmt.Printf("Error leyendo CSV: %v\n", err)
		return
	}

	header := records[0]
	datosOriginales := records[1:]
	conteoOriginal := len(datosOriginales)
	faltantes := objetivo - conteoOriginal

	fmt.Printf(" [INFO] Registros base: %d\n", conteoOriginal)
	fmt.Printf(" [INFO] Generando sinteticos: %d\n", faltantes)

	// Mapear columnas
	colIdx := make(map[string]int)
	for i, h := range header {
		colIdx[h] = i
	}

	// 2. Preparar Concurrencia (Paralelismo de Datos)
	numWorkers := runtime.NumCPU()
	registrosPorWorker := faltantes / numWorkers
	
	results := make(chan [][]string, numWorkers)
	var wg sync.WaitGroup

	fmt.Printf(" [PCD] Desplegando %d Goroutines para la generacion...\n", numWorkers)

	// 3. Lanzar Workers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		
		// El ultimo worker se lleva el residuo
		cantidad := registrosPorWorker
		if w == numWorkers-1 {
			cantidad = faltantes - (registrosPorWorker * w)
		}

		go func(workerID int, count int) {
			defer wg.Done()
			
			// Semilla local para cada goroutine para evitar patrones repetidos
			localRand := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))
			chunk := make([][]string, 0, count)

			for i := 0; i < count; i++ {
				// Elegir base al azar
				idxBase := localRand.Intn(conteoOriginal)
				originalRow := datosOriginales[idxBase]
				
				newRow := make([]string, len(originalRow))
				copy(newRow, originalRow)

				// RUIDO 1: ID Unico (WorkerID + Indice)
				if idx, ok := colIdx["ORDEN_ELECTRONICA"]; ok {
					newRow[idx] = fmt.Sprintf("%s-W%dI%d", originalRow[idx], workerID, i)
				}

				// RUIDO 2: Precios (+/- 2%)
				colsPrecios := []string{"TOTAL", "SUB_TOTAL", "IGV"}
				for _, cp := range colsPrecios {
					if idx, ok := colIdx[cp]; ok && newRow[idx] != "" {
						val, _ := strconv.ParseFloat(newRow[idx], 64)
						variacion := 0.98 + (localRand.Float64() * 0.04)
						newRow[idx] = strconv.FormatFloat(val*variacion, 'f', 2, 64)
					}
				}

				// RUIDO 3: Fechas (+/- 12h)
				colsFechas := []string{"FECHA_PROCESO", "FECHA_FORMALIZACION", "FECHA_ULTIMO_ESTADO"}
				for _, cf := range colsFechas {
					if idx, ok := colIdx[cf]; ok && newRow[idx] != "" {
						t, err := time.Parse("2006-01-02 15:04:05", newRow[idx])
						if err == nil {
							offset := time.Duration(localRand.Intn(24)-12) * time.Hour
							newRow[idx] = t.Add(offset).Format("2006-01-02 15:04:05")
						}
					}
				}
				chunk = append(chunk, newRow)
			}
			results <- chunk
		}(w, cantidad)
	}

	// 4. Crear archivo de salida y escribir
	outF, _ := os.Create(salida)
	defer outF.Close()
	writer := csv.NewWriter(outF)
	writer.Comma = ';'
	
	// Escribir cabecera y originales primero
	writer.Write(header)
	for _, row := range datosOriginales {
		writer.Write(row)
	}

	// 5. Reclectar resultados de workers
	go func() {
		wg.Wait()
		close(results)
	}()

	totalGen := 0
	for chunk := range results {
		for _, row := range chunk {
			writer.Write(row)
		}
		totalGen += len(chunk)
		fmt.Printf(" [OK] Lote de registros procesado y escrito en disco.\n")
	}
	writer.Flush()

	elapsed := time.Since(start)
	fmt.Println("============================================================")
	fmt.Printf(" PROCESO COMPLETADO EN: %v\n", elapsed)
	fmt.Printf(" Dataset de 1M guardado en: %s\n", salida)
	fmt.Println("============================================================")
}
