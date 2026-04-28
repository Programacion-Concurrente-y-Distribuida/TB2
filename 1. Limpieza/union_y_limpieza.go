package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Columnas canónicas de salida
var columnasEsperadas = []string{
	"FECHA_PUBLICACION", "OP", "ENTIDAD", "DISPOSITIVO",
	"NUMERO", "SUMILLA", "LINK", "FECHA_CORTE",
}

// Mapeo de variantes de cabecera (con tilde, espacios, mayúsculas) → nombre canónico
var aliasCol = map[string]string{
	"FECHA_PUBLICACION":  "FECHA_PUBLICACION",
	"FECHA PUBLICACION":  "FECHA_PUBLICACION",
	"FECHA_PUBLICACI_N":  "FECHA_PUBLICACION",
	"FECHA PUBLICACI_N":  "FECHA_PUBLICACION",
	"OP":                 "OP",
	"ENTIDAD":            "ENTIDAD",
	"DISPOSITIVO":        "DISPOSITIVO",
	"NUMERO":             "NUMERO",
	"N_MERO":             "NUMERO",
	"SUMILLA":            "SUMILLA",
	"LINK":               "LINK",
	"FECHA_CORTE":        "FECHA_CORTE",
	"FECHA CORTE":        "FECHA_CORTE",
}

func main() {
	start := time.Now()

	dataDir := filepath.Join("..", "data")
	outputFile := filepath.Join("..", "dataset_limpio.csv")

	fmt.Println("============================================================")
	fmt.Println(" Iniciando Union y Limpieza Secuencial - Dispositivos Legales")
	fmt.Println("============================================================")

	files, _ := filepath.Glob(filepath.Join(dataDir, "*.csv"))
	if len(files) == 0 {
		fmt.Println("Error: no se encontraron archivos .csv en", dataDir)
		return
	}
	sort.Strings(files)

	// Clave de deduplicación: OP (identificador único por dispositivo legal)
	seen := make(map[string]bool)
	var allRows [][]string
	totalLeidos := 0

	for _, f := range files {
		nombre := filepath.Base(f)
		rows, err := procesarArchivo(f, seen)
		if err != nil {
			fmt.Printf(" [ERR] %-50s | %v\n", nombre, err)
			continue
		}
		totalLeidos += len(rows)
		allRows = append(allRows, rows...)
		fmt.Printf(" [OK]  %-50s | +%d filas\n", nombre, len(rows))
	}

	fmt.Printf("\n Consolidados %d registros únicos. Guardando...\n", len(allRows))

	out, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Error creando archivo de salida:", err)
		return
	}
	defer out.Close()

	out.WriteString("\xef\xbb\xbf") // BOM para Excel
	w := csv.NewWriter(out)
	w.Comma = ','
	w.Write(columnasEsperadas)
	for _, row := range allRows {
		w.Write(row)
	}
	w.Flush()

	elapsed := time.Since(start)
	fmt.Println("============================================================")
	fmt.Printf(" PROCESO COMPLETADO EN: %v\n", elapsed)
	fmt.Printf(" Registros totales leídos : %d\n", totalLeidos)
	fmt.Printf(" Registros únicos guardados: %d\n", len(allRows))
	fmt.Printf(" Dataset guardado en: %s\n", outputFile)
	fmt.Println("============================================================")
}

func procesarArchivo(ruta string, seen map[string]bool) ([][]string, error) {
	raw, err := os.ReadFile(ruta)
	if err != nil {
		return nil, err
	}

	// Convertir Latin-1 a UTF-8 si hay bytes no-ASCII fuera de rango UTF-8 válido
	content := asegurarUTF8(raw)

	reader := csv.NewReader(bytes.NewReader(content))
	reader.Comma = ','
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, nil
	}

	// Mapear cabeceras → índice de columna
	idxPorCanon := make(map[string]int)
	for i, h := range records[0] {
		canon := canonicalizarCabecera(h)
		if canon != "" {
			idxPorCanon[canon] = i
		}
	}

	var resultado [][]string

	for _, rec := range records[1:] {
		fila := make([]string, len(columnasEsperadas))
		vacia := true

		for j, col := range columnasEsperadas {
			idx, ok := idxPorCanon[col]
			if !ok || idx >= len(rec) {
				continue
			}
			val := strings.TrimSpace(rec[idx])
			if val != "" {
				vacia = false
			}

			// Normalizar fechas a YYYY-MM-DD
			if col == "FECHA_PUBLICACION" || col == "FECHA_CORTE" {
				val = normalizarFecha(val)
			}
			fila[j] = val
		}

		if vacia {
			continue
		}

		// Descartar si falta fecha o OP
		if fila[0] == "" || fila[1] == "" {
			continue
		}

		// Deduplicar por OP
		op := fila[1]
		if seen[op] {
			continue
		}
		seen[op] = true

		resultado = append(resultado, fila)
	}

	return resultado, nil
}

// canonicalizarCabecera normaliza una cabecera a su nombre canónico.
func canonicalizarCabecera(h string) string {
	h = strings.TrimSpace(h)
	h = strings.ToUpper(h)

	// Reemplazar caracteres con tilde por equivalentes ASCII
	reemplazos := map[rune]rune{
		'Á': 'A', 'É': 'E', 'Í': 'I', 'Ó': 'O', 'Ú': 'U', 'Ñ': 'N',
		'á': 'A', 'é': 'E', 'í': 'I', 'ó': 'O', 'ú': 'U', 'ñ': 'N',
	}
	var sb strings.Builder
	for _, r := range h {
		if rep, ok := reemplazos[r]; ok {
			sb.WriteRune(rep)
		} else {
			sb.WriteRune(r)
		}
	}
	normalizado := sb.String()

	if canon, ok := aliasCol[normalizado]; ok {
		return canon
	}
	return ""
}

// normalizarFecha convierte YYYYMMDD o DD/MM/YYYY a YYYY-MM-DD.
func normalizarFecha(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	formatos := []string{
		"20060102",
		"2006-01-02",
		"02/01/2006",
		"2006/01/02",
		"2006-01-02 15:04:05",
	}
	for _, f := range formatos {
		if t, err := time.Parse(f, s); err == nil {
			return t.Format("2006-01-02")
		}
	}
	return s
}

// asegurarUTF8 convierte Latin-1 a UTF-8 cuando detecta bytes no válidos en UTF-8.
func asegurarUTF8(b []byte) []byte {
	if isUTF8(b) {
		return b
	}
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

func isUTF8(b []byte) bool {
	for i := 0; i < len(b); {
		r := b[i]
		if r < 0x80 {
			i++
			continue
		}
		var size int
		switch {
		case r&0xE0 == 0xC0:
			size = 2
		case r&0xF0 == 0xE0:
			size = 3
		case r&0xF8 == 0xF0:
			size = 4
		default:
			return false
		}
		if i+size > len(b) {
			return false
		}
		for j := 1; j < size; j++ {
			if b[i+j]&0xC0 != 0x80 {
				return false
			}
		}
		i += size
	}
	return true
}
