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

// ─────────────────────────────────────────────
// CONFIGURACIÓN
// ─────────────────────────────────────────────

type Job struct {
	URL      string
	Filename string
}

func main() {
	start := time.Now()
	baseURL := "https://www.datosabiertos.gob.pe/sites/default/files/ReportePCBienes"

	// 1. Detectar y crear carpeta "data"
	dataFolder := "data"
	if _, err := os.Stat(dataFolder); os.IsNotExist(err) {
		dataFolder = "../data"
	}
	os.MkdirAll(dataFolder, 0755)

	fmt.Println("============================================================")
	fmt.Println(" [PCD] Iniciando Descarga Concurrente de Datasets")
	fmt.Println("============================================================")

	// 2. Generar lista de URLs basadas en el patron ReportePCBienesYYYYMM.csv
	var jobsList []Job
	years := []int{2022, 2023, 2024, 2025, 2026}

	for _, year := range years {
		for month := 1; month <= 12; month++ {
			if year == 2026 && month > 3 {
				break
			}
			mesTexto := fmt.Sprintf("%02d", month)
			filename := fmt.Sprintf("ReportePCBienes%d%s.csv", year, mesTexto)
			fileURL := fmt.Sprintf("%s%d%s.csv", baseURL, year, mesTexto)
			jobsList = append(jobsList, Job{URL: fileURL, Filename: filename})
		}
	}

	// 3. Setup de Worker Pool
	// Usamos un numero controlado de workers (4) para no saturar al servidor del gobierno
	numWorkers := 4
	jobsChan := make(chan Job, len(jobsList))
	var wg sync.WaitGroup
	var mu sync.Mutex
	downloadedCount := 0

	// Lanzar Workers
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for job := range jobsChan {
				filePath := filepath.Join(dataFolder, job.Filename)
				fmt.Printf(" [*] Worker %d: Descargando %s\n", id, job.Filename)
				
				err := downloadFile(filePath, job.URL)
				if err != nil {
					// No imprimimos errores 404 como fallos criticos (meses no publicados)
					if err.Error() != "404 Not Found" {
						fmt.Printf(" [ERR] %s: %v\n", job.Filename, err)
					} else {
						fmt.Printf(" [!] %s no disponible en el servidor.\n", job.Filename)
					}
					os.Remove(filePath)
				} else {
					fmt.Printf(" [OK] %s guardado correctamente.\n", job.Filename)
					mu.Lock()
					downloadedCount++
					mu.Unlock()
				}
				// Pequena pausa para ser amigables con el servidor
				time.Sleep(200 * time.Millisecond)
			}
		}(w)
	}

	// Enviar trabajos al pool
	for _, j := range jobsList {
		jobsChan <- j
	}
	close(jobsChan)

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Println("============================================================")
	fmt.Printf(" PROCESO COMPLETADO EN: %v\n", elapsed)
	fmt.Printf(" Archivos descargados exitosamente: %d\n", downloadedCount)
	fmt.Println("============================================================")
}

func downloadFile(filepath string, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	// User-Agent real para evitar bloqueos por bots
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	return err
}
