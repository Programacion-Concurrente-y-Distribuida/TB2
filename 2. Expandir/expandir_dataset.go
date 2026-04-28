package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

func main() {
	start := time.Now()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	entrada := "../dataset_limpio.csv"
	salida := filepath.Join("..", "dataset_1m.csv")
	objetivo := 1_000_000

	fmt.Println("============================================================")
	fmt.Println(" Iniciando Expansion Secuencial a 1 Millon de Registros")
	fmt.Println("============================================================")

	// 1. Leer dataset limpio
	f, err := os.Open(entrada)
	if err != nil {
		fmt.Printf("Error abriendo entrada: %v\n", err)
		return
	}
	reader := csv.NewReader(f)
	reader.Comma = ','
	records, err := reader.ReadAll()
	f.Close()
	if err != nil {
		fmt.Printf("Error leyendo CSV: %v\n", err)
		return
	}
	if len(records) < 2 {
		fmt.Println("Error: dataset vacío o sin filas de datos.")
		return
	}

	header := records[0]
	base := records[1:]
	nBase := len(base)
	faltantes := objetivo - nBase

	fmt.Printf(" Registros base    : %d\n", nBase)
	fmt.Printf(" Registros sintéticos a generar: %d\n", faltantes)

	// Índices de columnas que se modifican para ruido
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

	outF.WriteString("\xef\xbb\xbf") // BOM para Excel
	writer := csv.NewWriter(outF)
	writer.Comma = ','

	// 3. Escribir cabecera y filas originales
	writer.Write(header)
	for _, row := range base {
		writer.Write(row)
	}

	// 4. Generar y escribir filas sintéticas de forma secuencial
	for i := 0; i < faltantes; i++ {
		original := base[rng.Intn(nBase)]
		nueva := make([]string, len(original))
		copy(nueva, original)

		// Ruido 1: OP único (base + sufijo sintético)
		if idx, ok := colIdx["OP"]; ok {
			nueva[idx] = fmt.Sprintf("%s-S%d", original[idx], i)
		}

		// Ruido 2: FECHA_PUBLICACION +/- hasta 7 días
		if idx, ok := colIdx["FECHA_PUBLICACION"]; ok && nueva[idx] != "" {
			if t, err := time.Parse("2006-01-02", nueva[idx]); err == nil {
				offset := time.Duration(rng.Intn(15)-7) * 24 * time.Hour
				nueva[idx] = t.Add(offset).Format("2006-01-02")
			}
		}

		// Ruido 3: FECHA_CORTE +/- hasta 7 días
		if idx, ok := colIdx["FECHA_CORTE"]; ok && nueva[idx] != "" {
			if t, err := time.Parse("2006-01-02", nueva[idx]); err == nil {
				offset := time.Duration(rng.Intn(15)-7) * 24 * time.Hour
				nueva[idx] = t.Add(offset).Format("2006-01-02")
			}
		}

		writer.Write(nueva)

		if (i+1)%100_000 == 0 {
			fmt.Printf(" [INFO] %d / %d registros sintéticos generados...\n", i+1, faltantes)
		}
	}

	writer.Flush()

	elapsed := time.Since(start)
	fmt.Println("============================================================")
	fmt.Printf(" PROCESO COMPLETADO EN: %v\n", elapsed)
	fmt.Printf(" Dataset de 1M guardado en: %s\n", salida)
	fmt.Println("============================================================")
}
