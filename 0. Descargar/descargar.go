package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func main() {
	// El archivo base a buscar tiene la estructura:
	// https://www.datosabiertos.gob.pe/sites/default/files/ReportePCBienesYYYYMM.csv
	baseURL := "https://www.datosabiertos.gob.pe/sites/default/files/ReportePCBienes"

	// Crear carpeta "data"
	dataFolder := "../data"
	err := os.MkdirAll(dataFolder, 0755)
	if err != nil {
		fmt.Printf("Error al crear carpeta 'data': %v\n", err)
		return
	}

	fmt.Println("[*] Iniciando descarga directa de los reportes mensuales de Órdenes de Compra...")

	var urlsToDownload []string
	years := []int{2022, 2023, 2024, 2025, 2026}

	for _, year := range years {
		for month := 1; month <= 12; month++ {
			if year == 2026 && month > 3 {
				break
			}
			mesTexto := fmt.Sprintf("%02d", month)
			filename := fmt.Sprintf("ReportePCBienes%d%s.csv", year, mesTexto)
			fileURL := fmt.Sprintf("%s%d%s.csv", baseURL, year, mesTexto)
			urlsToDownload = append(urlsToDownload, fileURL)
		}
	}

	var wg sync.WaitGroup
	// Limitamos concurrencia para evitar saturar el servidor y que nos rechacen
	sem := make(chan struct{}, 3)
	
	// Variables para contar resultados (usamos un mutex para sumar de forma segura)
	var downloadedCount int
	var mu sync.Mutex

	for idx, dURL := range urlsToDownload {
		wg.Add(1)
		sem <- struct{}{} // Ocupar un slot

		go func(url string, curIndex int) {
			defer wg.Done()
			defer func() { <-sem }()
			
			// Esperar ligeramente entre llamadas iniciales de goroutines
			time.Sleep(time.Duration(curIndex*100) * time.Millisecond)

			filename := filepath.Base(url)
			filePath := filepath.Join(dataFolder, filename)

			fmt.Printf("[*] Buscando: %s\n", filename)
			
			err := downloadCSV(filePath, url)
			if err != nil {
				// Evitamos imprimir 404 como un error catastrófico, es posible que no exista ese mes
				if err.Error() == "código HTTP 404 Not Found" {
					fmt.Printf("  -> No encontrado: %s no publicado.\n", filename)
				} else {
					fmt.Printf("  -> Error en %s: %v\n", filename, err)
				}
				// Si falla o no existe, borramos el archivo en blanco creado
				os.Remove(filePath)
			} else {
				fmt.Printf("  -> ¡Descargado!: %s\n", filename)
				mu.Lock()
				downloadedCount++
				mu.Unlock()
			}
		}(dURL, idx)
	}

	wg.Wait()
	fmt.Println("==================================================")
	fmt.Printf("[Terminado] Proceso completado. Se descargaron exitosamente %d archivos de %d posibles.\n", downloadedCount, len(urlsToDownload))
}

func downloadCSV(filepath string, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	client := &http.Client{
		Timeout: 30 * time.Second, // Timeout para evitar bloqueos
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("código HTTP %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	return err
}
