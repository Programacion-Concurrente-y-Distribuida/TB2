// union_y_limpieza.go
// -------------------
// Une todos los .csv de la carpeta /data y aplica limpieza de datos
// sobre el dataset de Ordenes de Compra - Catalogos Electronicos (PERU COMPRAS).
//
// Salida: data/dataset_limpio.csv
// Ejecutar: go run union_y_limpieza.go

package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ─────────────────────────────────────────────
// COLUMNAS OFICIALES DEL DATASET
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

// ─────────────────────────────────────────────
// UTILIDADES
// ─────────────────────────────────────────────

// latin1ToUTF8 convierte bytes latin-1 a UTF-8.
// Latin-1 = Unicode 0-255, cada byte > 127 se codifica en 2 bytes UTF-8.
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

// normalizarCol estandariza un nombre de columna a ASCII puro sin tildes.
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
		// caracteres no-ASCII no mapeados se descartan
	}
	result := sb.String()
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}
	return result
}

// esNulo retorna true si el valor está vacío o solo tiene espacios.
func esNulo(s string) bool {
	return strings.TrimSpace(s) == ""
}

// parsearFecha intenta convertir un string a time.Time con varios formatos.
func parsearFecha(s string) (time.Time, bool) {
	if esNulo(s) {
		return time.Time{}, false
	}
	formatos := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
		"02/01/2006",
		"2006/01/02",
	}
	s = strings.TrimSpace(s)
	for _, f := range formatos {
		if t, err := time.Parse(f, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// parsearFloat limpia y convierte un string a float64.
func parsearFloat(s string) (float64, bool) {
	if esNulo(s) {
		return 0, false
	}
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", ".")
	s = strings.ReplaceAll(s, " ", "")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// formatearInt formatea un entero con separador de miles (coma).
func formatearInt(n int) string {
	s := strconv.Itoa(n)
	neg := false
	if n < 0 {
		s = strconv.Itoa(-n)
		neg = true
	}
	// insertar comas
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	if neg {
		return "-" + string(result)
	}
	return string(result)
}

// leerCSV lee un archivo CSV (latin-1) y devuelve headers normalizados + filas
// filtradas a las columnas esperadas.
func leerCSV(ruta string) ([]string, [][]string, error) {
	rawBytes, err := os.ReadFile(ruta)
	if err != nil {
		return nil, nil, err
	}

	// Convertir de latin-1 a UTF-8
	utf8Bytes := latin1ToUTF8(rawBytes)

	reader := csv.NewReader(bytes.NewReader(utf8Bytes))
	reader.Comma = ';'
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(records) == 0 {
		return nil, nil, nil
	}

	// Normalizar cabecera
	rawHeaders := records[0]
	headers := make([]string, len(rawHeaders))
	for i, h := range rawHeaders {
		headers[i] = normalizarCol(h)
	}

	// Mapear columnas disponibles
	colIdx := make(map[string]int)
	for i, h := range headers {
		colIdx[h] = i
	}

	// Filtrar a columnas esperadas
	var filteredCols []string
	var filteredIdx []int
	for _, ec := range columnasEsperadas {
		if idx, ok := colIdx[ec]; ok {
			filteredCols = append(filteredCols, ec)
			filteredIdx = append(filteredIdx, idx)
		}
	}

	// Construir filas
	rows := make([][]string, 0, len(records)-1)
	for _, record := range records[1:] {
		row := make([]string, len(filteredCols))
		for i, idx := range filteredIdx {
			if idx < len(record) {
				row[i] = record[idx]
			}
		}
		rows = append(rows, row)
	}

	return filteredCols, rows, nil
}

// ─────────────────────────────────────────────
// MAIN
// ─────────────────────────────────────────────
func main() {
	baseDir, _ := os.Getwd()
	dataDir := filepath.Join(baseDir, "data")
	salida := filepath.Join(dataDir, "dataset_limpio.csv")
	sep := strings.Repeat("=", 60)

	// ──────────────────────────────────────────────────────────
	// PASO 1 - Union de todos los CSV
	// ──────────────────────────────────────────────────────────
	fmt.Println(sep)
	fmt.Println("  PASO 1 - Leyendo y uniendo todos los archivos CSV")
	fmt.Println(sep)

	entries, err := os.ReadDir(dataDir)
	if err != nil {
		fmt.Printf("Error leyendo directorio: %v\n", err)
		os.Exit(1)
	}

	var csvFiles []string
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() && strings.ToLower(filepath.Ext(name)) == ".csv" && name != "dataset_limpio.csv" {
			csvFiles = append(csvFiles, filepath.Join(dataDir, name))
		}
	}
	sort.Strings(csvFiles)

	if len(csvFiles) == 0 {
		fmt.Printf("No se encontraron archivos .csv en: %s\n", dataDir)
		os.Exit(1)
	}

	var allHeaders []string
	var allRows [][]string
	headerIdx := make(map[string]int)

	for _, csvPath := range csvFiles {
		nombre := filepath.Base(csvPath)
		headers, rows, err := leerCSV(csvPath)
		if err != nil {
			fmt.Printf("  [ERR] %-40s - ERROR: %v\n", nombre, err)
			continue
		}

		// Primera vez: establecer cabecera maestra
		if allHeaders == nil {
			allHeaders = headers
			for i, h := range allHeaders {
				headerIdx[h] = i
			}
		}

		// Construir índice local del archivo
		localIdx := make(map[string]int)
		for i, h := range headers {
			localIdx[h] = i
		}

		// Alinear filas a cabecera maestra
		for _, row := range rows {
			newRow := make([]string, len(allHeaders))
			for i, h := range allHeaders {
				if idx, ok := localIdx[h]; ok && idx < len(row) {
					newRow[i] = row[idx]
				}
			}
			allRows = append(allRows, newRow)
		}

		fmt.Printf("  [OK] %-40s %8s filas\n", nombre, formatearInt(len(rows)))
	}

	fmt.Printf("\n  Total tras union: %s filas  |  %d columnas\n\n",
		formatearInt(len(allRows)), len(allHeaders))

	// ──────────────────────────────────────────────────────────
	// PASO 2 - Diagnostico inicial
	// ──────────────────────────────────────────────────────────
	fmt.Println(sep)
	fmt.Println("  PASO 2 - Diagnostico inicial")
	fmt.Println(sep)
	fmt.Println("\n[Columnas y valores nulos]")

	totalRows := len(allRows)
	nullCounts := make([]int, len(allHeaders))
	for _, row := range allRows {
		for i, val := range row {
			if esNulo(val) {
				nullCounts[i]++
			}
		}
	}
	for i, h := range allHeaders {
		pct := float64(nullCounts[i]) / float64(totalRows) * 100
		fmt.Printf("  %-40s nulos: %8s  (%5.1f%%)\n", h, formatearInt(nullCounts[i]), pct)
	}

	// ──────────────────────────────────────────────────────────
	// PASO 3 - Limpieza de datos
	// ──────────────────────────────────────────────────────────
	fmt.Printf("\n%s\n", sep)
	fmt.Println("  PASO 3 - Limpieza de datos")
	fmt.Println(sep)

	filasInicial := len(allRows)

	// 3a. Eliminar columnas 100% vacias
	antCols := len(allHeaders)
	var newHeaders []string
	var keepCols []int
	for i, h := range allHeaders {
		if nullCounts[i] < totalRows {
			newHeaders = append(newHeaders, h)
			keepCols = append(keepCols, i)
		}
	}
	filtered := make([][]string, len(allRows))
	for r, row := range allRows {
		newRow := make([]string, len(newHeaders))
		for ni, ci := range keepCols {
			if ci < len(row) {
				newRow[ni] = row[ci]
			}
		}
		filtered[r] = newRow
	}
	allRows = filtered
	allHeaders = newHeaders
	// Recalcular headerIdx
	headerIdx = make(map[string]int)
	for i, h := range allHeaders {
		headerIdx[h] = i
	}
	fmt.Printf("\n  [3a] Columnas 100%% vacias eliminadas: %d\n", antCols-len(allHeaders))
	fmt.Printf("       Columnas restantes: %v\n", allHeaders)

	// 3b. Eliminar duplicados
	antes := len(allRows)
	seen := make(map[string]bool, len(allRows))
	var dedupRows [][]string
	for _, row := range allRows {
		key := strings.Join(row, "\x00")
		if !seen[key] {
			seen[key] = true
			dedupRows = append(dedupRows, row)
		}
	}
	allRows = dedupRows
	fmt.Printf("\n  [3b] Duplicados eliminados: %s\n", formatearInt(antes-len(allRows)))

	// 3c. Eliminar filas totalmente vacias
	antes = len(allRows)
	var nonEmpty [][]string
	for _, row := range allRows {
		allNull := true
		for _, v := range row {
			if !esNulo(v) {
				allNull = false
				break
			}
		}
		if !allNull {
			nonEmpty = append(nonEmpty, row)
		}
	}
	allRows = nonEmpty
	fmt.Printf("  [3c] Filas completamente vacias eliminadas: %s\n", formatearInt(antes-len(allRows)))

	// 3d. Convertir columnas de fecha a datetime
	var fechaCols []string
	for _, h := range allHeaders {
		if strings.Contains(h, "FECHA") {
			fechaCols = append(fechaCols, h)
		}
	}
	for _, col := range fechaCols {
		idx := headerIdx[col]
		for r, row := range allRows {
			if !esNulo(row[idx]) {
				if t, ok := parsearFecha(row[idx]); ok {
					allRows[r][idx] = t.Format("2006-01-02 15:04:05")
				} else {
					allRows[r][idx] = ""
				}
			}
		}
	}
	fmt.Printf("\n  [3d] Fechas convertidas a datetime: %v\n", fechaCols)

	// 3e. Convertir columnas numericas
	numColNames := []string{"SUB_TOTAL", "IGV", "TOTAL"}
	var numCols []string
	for _, c := range numColNames {
		if _, ok := headerIdx[c]; ok {
			numCols = append(numCols, c)
		}
	}
	for _, col := range numCols {
		idx := headerIdx[col]
		for r, row := range allRows {
			if !esNulo(row[idx]) {
				if v, ok := parsearFloat(row[idx]); ok {
					allRows[r][idx] = strconv.FormatFloat(v, 'f', 2, 64)
				} else {
					allRows[r][idx] = ""
				}
			}
		}
	}
	fmt.Printf("  [3e] Columnas convertidas a numerico: %v\n", numCols)

	// 3f. Eliminar filas con TOTAL nulo o <= 0
	antes = len(allRows)
	if tidx, ok := headerIdx["TOTAL"]; ok {
		var validRows [][]string
		for _, row := range allRows {
			if !esNulo(row[tidx]) {
				if v, err2 := strconv.ParseFloat(row[tidx], 64); err2 == nil && v > 0 {
					validRows = append(validRows, row)
				}
			}
		}
		allRows = validRows
	}
	fmt.Printf("  [3f] Filas con TOTAL invalido (nulo/<=0): %s\n", formatearInt(antes-len(allRows)))

	// 3g. Strip de espacios en columnas de texto
	numericSet := map[string]bool{"SUB_TOTAL": true, "IGV": true, "TOTAL": true}
	dateSet := map[string]bool{}
	for _, c := range fechaCols {
		dateSet[c] = true
	}
	textCount := 0
	for _, h := range allHeaders {
		if !numericSet[h] && !dateSet[h] {
			textCount++
		}
	}
	for r, row := range allRows {
		for i, h := range allHeaders {
			if !numericSet[h] && !dateSet[h] {
				allRows[r][i] = strings.TrimSpace(row[i])
			}
		}
	}
	fmt.Printf("  [3g] Strip aplicado a %d columnas de texto\n", textCount)

	// 3h. Normalizar RUCs a 11 digitos
	for r, row := range allRows {
		for i, h := range allHeaders {
			if strings.Contains(h, "RUC") {
				ruc := strings.TrimSpace(row[i])
				for len(ruc) < 11 {
					ruc = "0" + ruc
				}
				allRows[r][i] = ruc
			}
		}
	}
	fmt.Println("  [3h] RUCs normalizados a 11 digitos")

	// 3i. Crear columna ANO_MES
	if fidx, ok := headerIdx["FECHA_FORMALIZACION"]; ok {
		allHeaders = append(allHeaders, "ANO_MES")
		for r, row := range allRows {
			anoMes := ""
			if !esNulo(row[fidx]) {
				if t, ok2 := parsearFecha(row[fidx]); ok2 {
					anoMes = t.Format("2006-01")
				}
			}
			allRows[r] = append(allRows[r], anoMes)
		}
		fmt.Println("  [3i] Columna ANO_MES creada desde FECHA_FORMALIZACION")
	}

	// ──────────────────────────────────────────────────────────
	// PASO 4 - Resumen final
	// ──────────────────────────────────────────────────────────
	fmt.Printf("\n%s\n", sep)
	fmt.Println("  PASO 4 - Resumen final")
	fmt.Println(sep)
	fmt.Printf("\n  Filas antes de la limpieza :  %10s\n", formatearInt(filasInicial))
	fmt.Printf("  Filas despues de la limpieza: %10s\n", formatearInt(len(allRows)))
	fmt.Printf("  Filas eliminadas (total)    : %10s\n", formatearInt(filasInicial-len(allRows)))
	fmt.Printf("  Columnas finales            : %d\n", len(allHeaders))

	fmt.Println("\n  Columnas del dataset final:")
	finalNulls := make([]int, len(allHeaders))
	for _, row := range allRows {
		for i, v := range row {
			if esNulo(v) {
				finalNulls[i]++
			}
		}
	}
	for i, h := range allHeaders {
		fmt.Printf("    %-40s nulos: %s\n", h, formatearInt(finalNulls[i]))
	}

	// ──────────────────────────────────────────────────────────
	// PASO 5 - Guardar resultado
	// ──────────────────────────────────────────────────────────
	fmt.Printf("\n%s\n", sep)
	fmt.Println("  PASO 5 - Guardando dataset limpio")
	fmt.Println(sep)

	outFile, err := os.Create(salida)
	if err != nil {
		fmt.Printf("Error creando archivo de salida: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	// BOM UTF-8 para compatibilidad con Excel
	outFile.WriteString("\xef\xbb\xbf")

	writer := csv.NewWriter(outFile)
	writer.Comma = ';'
	writer.Write(allHeaders)
	for _, row := range allRows {
		writer.Write(row)
	}
	writer.Flush()

	info, _ := os.Stat(salida)
	tamMB := float64(info.Size()) / (1024 * 1024)
	fmt.Printf("\n  [OK] Archivo guardado en: %s\n", salida)
	fmt.Printf("       Tamano: %.2f MB  |  Filas: %s\n\n", tamMB, formatearInt(len(allRows)))
}
