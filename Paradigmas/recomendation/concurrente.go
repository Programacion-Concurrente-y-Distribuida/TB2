package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"sync"
)

func main() {
	inputPath := "../../dataset_limpio.csv"
	outputPath := "../../dataset_lematizado.csv"
	stopwordsPath := "../../data/stopwords_es.txt"
	lemmasPath := "../../data/lemmas_es.txt"

	if len(os.Args) > 1 {
		inputPath = os.Args[1]
	}
	if len(os.Args) > 2 {
		outputPath = os.Args[2]
	}

	numWorkers := runtime.NumCPU()
	if len(os.Args) > 3 {
		n, err := strconv.Atoi(os.Args[3])
		if err != nil || n <= 0 {
			panic("el numero de workers debe ser un entero mayor a 0")
		}
		numWorkers = n
	}

	stopwords, err := loadStopwords(stopwordsPath)
	if err != nil {
		panic(fmt.Errorf("cargar stopwords: %w", err))
	}
	lemmas, err := loadLemmas(lemmasPath)
	if err != nil {
		panic(fmt.Errorf("cargar lemas: %w", err))
	}

	in, err := os.Open(inputPath)
	if err != nil {
		panic(err)
	}
	defer in.Close()

	out, err := mustOpenWriter(outputPath)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	reader := csv.NewReader(in)
	writer := csv.NewWriter(out)
	defer writer.Flush()

	header, err := reader.Read()
	if err != nil {
		panic(err)
	}
	if err := writer.Write(withoutLinkColumn(header)); err != nil {
		panic(err)
	}

	jobs := make(chan RowJob, 2000)
	results := make(chan RowJob, 2000)

	var workers sync.WaitGroup
	workers.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer workers.Done()
			for job := range jobs {
				processRow(job.Row, stopwords, lemmas)
				results <- job
			}
		}()
	}

	go func() {
		workers.Wait()
		close(results)
	}()

	var feeder sync.WaitGroup
	feeder.Add(1)
	go func() {
		defer feeder.Done()
		idx := 0
		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				continue
			}
			jobs <- RowJob{Index: idx, Row: append([]string(nil), row...)}
			idx++
		}
		close(jobs)
	}()

	pending := make(map[int][]string)
	nextExpected := 0
	processed := 0
	for r := range results {
		pending[r.Index] = r.Row
		for {
			row, ok := pending[nextExpected]
			if !ok {
				break
			}
			delete(pending, nextExpected)
			if err := writer.Write(withoutLinkColumn(row)); err != nil {
				panic(err)
			}
			nextExpected++
			processed++
		}
	}

	feeder.Wait()
	fmt.Printf("filas procesadas: %d (workers=%d)\n", processed, numWorkers)
}
