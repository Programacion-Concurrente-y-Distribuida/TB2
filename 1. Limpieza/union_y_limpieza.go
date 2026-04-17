package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ─────────────────────────────────────────────
// CONFIGURACIÓN Y COLUMNAS
// ─────────────────────────────────────────────
var columnasEsperadas = []string{
	"FECHA_PROCESO", "RUC_PROVEEDOR", "PROVEEDOR",
	"RUC_ENTIDAD", "ENTIDAD", "TIPO_PROCEDIMIENTO",
	"ORDEN_ELECTRONICA", "ORDEN_ELECTRONICA_GENERADA",
	"ESTADO_ORDEN_ELECTRONICA", "DOCUMENTO_ESTADO_OCAM",
	"FECHA_FORMALIZACION", "FECHA_ULTIMO_ESTADO",
	"SUB_TOTAL", "IGV", "TOTAL",
	"ORDEN_DIGITALIZADA", "DESCRIPCION_ESTADO",
	"DESCRIPCION_CESION_DERECHOS", "ACUERDO_MARCO",
}

// Estructura para pasar trabajos a los workers
type Job struct {
	Path string
}

// Estructura para recibir resultados de los workers
type Result struct {
	Rows  [][]string
	Error error
	File  string
}

// ─────────────────────────────────────────────
// UTILIDADES DE LIMPIEZA
// ─────────────────────────────────────────────

func latin1ToUTF8(b []byte) []byte {
	var buf bytes.Buffer
	for _, c := range b {
		if c < 128 {
			buf.WriteByte(c)
		} else {
			buf.WriteByte(0xC0 | (c >> 6))
			buf.WriteByte(0x80 | (c & 0x3F))
		}
	}
	return buf.Bytes()
}

func normalizarCol(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)
	reemplazos := map[rune]string{
		'Ó': "O", 'É': "E", 'Í': "I", 'Á': "A", 'Ú': "U", 'Ñ': "N",
		'ó': "O", 'é': "E", 'í': "I", 'á': "A", 'ú': "U", 'ñ': "N",
		' ': "_",
	}
	var sb strings.Builder
	for _, r := range s {
		if rep, ok := reemplazos[r]; ok {
			sb.WriteString(rep)
		} else if r <= 127 {
			sb.WriteRune(r)
		}
	}
	res := sb.String()
	for strings.Contains(res, "__") {
		res = strings.ReplaceAll(res, "__", "_")
	}
	return res
}

func parsearFecha(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	formatos := []string{"2006-01-02 15:04:05", "2006-01-02", "02/01/2006", "2006/01/02"}
	for _, f := range formatos {
		if t, err := time.Parse(f, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func parsearFloat(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	s = strings.ReplaceAll(s, ",", ".")
	s = strings.ReplaceAll(s, " ", "")
	v, err := strconv.ParseFloat(s, 64)
	return v, err == nil
}

// ─────────────────────────────────────────────
// WORKER LOGIC
// ─────────────────────────────────────────────

func worker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		rows, err := procesarArchivo(job.Path)
		results <- Result{Rows: rows, Error: err, File: filepath.Base(job.Path)}
	}
}

func procesarArchivo(ruta string) ([][]string, error) {
	f, err := os.Open(ruta)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Leer todo y convertir a UTF-8
	content, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	utf8Content := latin1ToUTF8(content)

	reader := csv.NewReader(bytes.NewReader(utf8Content))
	reader.Comma = ';'
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, nil
	}

	// Identificar indices de columnas
	headers := records[0]
	colIdxMap := make(map[string]int)
	for i, h := range headers {
		colIdxMap[normalizarCol(h)] = i
	}

	var cleanedRows [][]string
	for _, record := range records[1:] {
		// 1. Filtrar solo las 19 columnas esperadas + 1 para ANO_MES
		newRow := make([]string, len(columnasEsperadas)+1)
		
		// 2. Limpieza de campos
		isValid := true
		var fechaFormalizacion time.Time

		for i, colName := range columnasEsperadas {
			originalIdx, exists := colIdxMap[colName]
			val := ""
			if exists && originalIdx < len(record) {
				val = strings.TrimSpace(record[originalIdx])
			}

			// Transformaciones específicas
			switch colName {
			case "TOTAL":
				if v, ok := parsearFloat(val); ok && v > 0 {
					val = strconv.FormatFloat(v, 'f', 2, 64)
				} else {
					isValid = false // Descartar si el total es inválido o <= 0
				}
			case "SUB_TOTAL", "IGV":
				if v, ok := parsearFloat(val); ok {
					val = strconv.FormatFloat(v, 'f', 2, 64)
				}
			case "FECHA_FORMALIZACION":
				if t, ok := parsearFecha(val); ok {
					val = t.Format("2006-01-02 15:04:05")
					fechaFormalizacion = t
				}
			case "FECHA_PROCESO", "FECHA_ULTIMO_ESTADO":
				if t, ok := parsearFecha(val); ok {
					val = t.Format("2006-01-02 15:04:05")
				}
			}

			// Normalizar RUC a 11 dígitos
			if strings.Contains(colName, "RUC") && val != "" {
				for len(val) < 11 {
					val = "0" + val
				}
			}

			newRow[i] = val
		}

		if isValid {
			// Crear columna ANO_MES (en la última posición)
			if !fechaFormalizacion.IsZero() {
				newRow[len(columnasEsperadas)] = fechaFormalizacion.Format("2006-01")
			}
			cleanedRows = append(cleanedRows, newRow)
		}
	}

	return cleanedRows, nil
}

// ─────────────────────────────────────────────
// MAIN
// ─────────────────────────────────────────────

func main() {
	start := time.Now()
	
	// Configuración de concurrencia: puedes ajustar este número para tus pruebas de estrés
	// El valor recomendado es runtime.NumCPU()
	numWorkers := runtime.NumCPU()
	
	fmt.Println("============================================================")
	fmt.Printf(" [PCD] Iniciando Limpieza Concurrente con %d Workers\n", numWorkers)
	fmt.Println("============================================================")

	dataDir := "data"
	// Si no existe 'data' en el directorio actual, probamos un nivel arriba
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		dataDir = "../data"
	}
	
	outputFile := filepath.Join(filepath.Dir(dataDir), "dataset_limpio.csv")

	files, _ := filepath.Glob(filepath.Join(dataDir, "*.csv"))
	if len(files) == 0 {
		fmt.Println("Error: No se encontraron archivos .csv en", dataDir)
		return
	}

	jobs := make(chan Job, len(files))
	results := make(chan Result, len(files))
	var wg sync.WaitGroup

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

	// 3. Recolectar resultados en paralelo con el procesamiento
	var finalDataset [][]string
	var mu sync.Mutex
	var totalRecords int
	
	// Goroutine para esperar y cerrar resultados
	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		if res.Error != nil {
			fmt.Printf(" [ERR] %s: %v\n", res.File, res.Error)
			continue
		}
		mu.Lock()
		finalDataset = append(finalDataset, res.Rows...)
		totalRecords += len(res.Rows)
		mu.Unlock()
		fmt.Printf(" [OK] %-35s | +%d filas\n", res.File, len(res.Rows))
	}

	// 4. Guardar resultado final
	fmt.Println("\n------------------------------------------------------------")
	fmt.Printf(" Consolidados %d registros. Guardando...\n", totalRecords)

	out, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Error creando archivo de salida:", err)
		return
	}
	defer out.Close()

	out.WriteString("\xef\xbb\xbf") // BOM para Excel
	writer := csv.NewWriter(out)
	writer.Comma = ';'
	
	// Escribir cabecera (oficiales + ANO_MES)
	header := append(columnasEsperadas, "ANO_MES")
	writer.Write(header)
	
	for _, row := range finalDataset {
		writer.Write(row)
	}
	writer.Flush()

	elapsed := time.Since(start)
	fmt.Println("============================================================")
	fmt.Printf(" PROCESO COMPLETADO EN: %v\n", elapsed)
	fmt.Printf(" Dataset guardado en: %s\n", outputFile)
	fmt.Println("============================================================")
}
