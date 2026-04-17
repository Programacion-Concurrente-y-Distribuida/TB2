package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	dataDir := filepath.Join("..", "data")
	entrada := filepath.Join(dataDir, "dataset_limpio.csv")
	salida := filepath.Join(dataDir, "dataset_1m.csv")
	objetivo := 1000000

	fmt.Println("--- Expandiendo Dataset a 1 Millon de Registros ---")

	// 1. Leer dataset original
	f, err := os.Open(entrada)
	if err != nil {
		fmt.Printf("Error abriendo entrada: %v\n", err)
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = ';'
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Error leyendo CSV: %v\n", err)
		return
	}

	header := records[0]
	datosOriginales := records[1:]
	conteoOriginal := len(datosOriginales)
	faltantes := objetivo - conteoOriginal

	fmt.Printf("Registros actuales: %d\n", conteoOriginal)
	fmt.Printf("Generando sinteticos: %d\n", faltantes)

	// Mapear columnas para saber que modificar
	colIdx := make(map[string]int)
	for i, h := range header {
		colIdx[h] = i
	}

	// 2. Crear archivo de salida
	outF, err := os.Create(salida)
	if err != nil {
		fmt.Printf("Error creando salida: %v\n", err)
		return
	}
	defer outF.Close()

	writer := csv.NewWriter(outF)
	writer.Comma = ';'
	defer writer.Flush()

	// Escribir cabecera
	writer.Write(header)

	// Escribir originales
	for _, row := range datosOriginales {
		writer.Write(row)
	}

	// 3. Generar sinteticos con ruido
	for i := 1; i <= faltantes; i++ {
		// Elegir un registro base al azar
		idxRand := rand.Intn(conteoOriginal)
		originalRow := datosOriginales[idxRand]
		
		// Clonar la fila
		newRow := make([]string, len(originalRow))
		copy(newRow, originalRow)

		// RUIDO 1: Identificador de Orden (hacerlo unico)
		if idx, ok := colIdx["ORDEN_ELECTRONICA"]; ok {
			newRow[idx] = fmt.Sprintf("%s-S%d", originalRow[idx], i)
		}

		// RUIDO 2: Precios (variacion aleatoria de +/- 2%)
		colsPrecios := []string{"TOTAL", "SUB_TOTAL", "IGV"}
		for _, cp := range colsPrecios {
			if idx, ok := colIdx[cp]; ok && newRow[idx] != "" {
				val, _ := strconv.ParseFloat(newRow[idx], 64)
				variacion := 1.0 + (rand.Float64()*0.04 - 0.02) // entre 0.98 y 1.02
				newRow[idx] = strconv.FormatFloat(val*variacion, 'f', 2, 64)
			}
		}

		// RUIDO 3: Fechas (variacion de +/- 12 horas)
		colsFechas := []string{"FECHA_PROCESO", "FECHA_FORMALIZACION", "FECHA_ULTIMO_ESTADO"}
		for _, cf := range colsFechas {
			if idx, ok := colIdx[cf]; ok && newRow[idx] != "" {
				t, err := time.Parse("2006-01-02 15:04:05", newRow[idx])
				if err == nil {
					randomHours := time.Duration(rand.Intn(24)-12) * time.Hour
					newRow[idx] = t.Add(randomHours).Format("2006-01-02 15:04:05")
				}
			}
		}

		writer.Write(newRow)

		if i%100000 == 0 {
			fmt.Printf("... %d registros generados\n", i)
		}
	}

	fmt.Println("--------------------------------------------------")
	fmt.Printf("Listo! Dataset final guardado en: %s\n", salida)
	fmt.Printf("Total de filas: %d\n", objetivo)
}
