package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
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

	processed := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		processRow(row, stopwords, lemmas)
		if err := writer.Write(withoutLinkColumn(row)); err != nil {
			panic(err)
		}
		processed++
	}

	fmt.Printf("filas procesadas: %d\n", processed)
}
