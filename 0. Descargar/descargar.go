package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Dataset struct {
	URL      string
	Filename string
}

var datasets = []Dataset{
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20130101_20220331.CSV",
		Filename: "DatosAbiertos_Periodo_20130101_20220331.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20220101_20221231.CSV",
		Filename: "DatosAbiertos_Periodo_20220101_20221231.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20230101_20231231.CSV",
		Filename: "DatosAbiertos_Periodo_20230101_20231231.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20230101_20230131.CSV",
		Filename: "DatosAbiertos_Periodo_20230101_20230131.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20230201_20230228.CSV",
		Filename: "DatosAbiertos_Periodo_20230201_20230228.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20230301_20230331.CSV",
		Filename: "DatosAbiertos_Periodo_20230301_20230331.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20230401_20230430.CSV",
		Filename: "DatosAbiertos_Periodo_20230401_20230430.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20230501_20230531.CSV",
		Filename: "DatosAbiertos_Periodo_20230501_20230531.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20230601_20230630.CSV",
		Filename: "DatosAbiertos_Periodo_20230601_20230630.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20230701_20230731.CSV",
		Filename: "DatosAbiertos_Periodo_20230701_20230731.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20230801_20230831.CSV",
		Filename: "DatosAbiertos_Periodo_20230801_20230831.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20230901_20230930.CSV",
		Filename: "DatosAbiertos_Periodo_20230901_20230930.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20231001_20231031.CSV",
		Filename: "DatosAbiertos_Periodo_20231001_20231031.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20231101_20231130.CSV",
		Filename: "DatosAbiertos_Periodo_20231101_20231130.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20231201_20231231.CSV",
		Filename: "DatosAbiertos_Periodo_20231201_20231231.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20240101_20240131.CSV",
		Filename: "DatosAbiertos_Periodo_20240101_20240131.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20240201_20240229.CSV",
		Filename: "DatosAbiertos_Periodo_20240201_20240229.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20240301_20240331.CSV",
		Filename: "DatosAbiertos_Periodo_20240301_20240331.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20240401_20240430.CSV",
		Filename: "DatosAbiertos_Periodo_20240401_20240430.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20240501_20240531.CSV",
		Filename: "DatosAbiertos_Periodo_20240501_20240531.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20240601_20240630.CSV",
		Filename: "DatosAbiertos_Periodo_20240601_20240630.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20240701_20240731.CSV",
		Filename: "DatosAbiertos_Periodo_20240701_20240731.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20240801_20240831.CSV",
		Filename: "DatosAbiertos_Periodo_20240801_20240831.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20240901_20240930.CSV",
		Filename: "DatosAbiertos_Periodo_20240901_20240930.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20241001_20241031.CSV",
		Filename: "DatosAbiertos_Periodo_20241001_20241031.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20241101_20241130.csv",
		Filename: "DatosAbiertos_Periodo_20241101_20241130.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20241201_20241231.csv",
		Filename: "DatosAbiertos_Periodo_20241201_20241231.csv",
	},
	{
		URL:      "https://www.datosabiertos.gob.pe/sites/default/files/DatosAbiertos_Periodo_20250101_20250131.csv",
		Filename: "DatosAbiertos_Periodo_20250101_20250131.csv",
	},
}

func main() {
	start := time.Now()

	dataFolder := filepath.Join("..", "data")
	os.MkdirAll(dataFolder, 0755)

	fmt.Println("============================================================")
	fmt.Println(" Iniciando Descarga Secuencial - Dispositivos Legales")
	fmt.Printf(" Total de archivos a descargar: %d\n", len(datasets))
	fmt.Println("============================================================")

	downloadedCount := 0

	for i, ds := range datasets {
		filePath := filepath.Join(dataFolder, ds.Filename)
		fmt.Printf(" [%d/%d] Descargando %s\n", i+1, len(datasets), ds.Filename)

		err := downloadFile(filePath, ds.URL)
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				fmt.Printf(" [!] No disponible en el servidor.\n")
			} else {
				fmt.Printf(" [ERR] %v\n", err)
			}
			os.Remove(filePath)
		} else {
			fmt.Printf(" [OK] Guardado correctamente.\n")
			downloadedCount++
		}

		time.Sleep(200 * time.Millisecond)
	}

	elapsed := time.Since(start)
	fmt.Println("============================================================")
	fmt.Printf(" PROCESO COMPLETADO EN: %v\n", elapsed)
	fmt.Printf(" Archivos descargados exitosamente: %d / %d\n", downloadedCount, len(datasets))
	fmt.Println("============================================================")
}

func downloadFile(filePath string, url string) error {
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	client := &http.Client{Timeout: 120 * time.Second}
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
