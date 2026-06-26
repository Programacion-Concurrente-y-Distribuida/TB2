// Comando predict-linear: carga un modelo de regresion lineal YA ENTRENADO
// (guardado con `train-linear -save-model modelo.json`) y lo aplica sobre un
// CSV de entrada SIN volver a entrenar.
//
// Uso minimo:
//
//	go run ./cmd/predict-linear -model modelo.json -input aqs_clean.csv
package main

import (
	"aqsml/internal/aqsml"
	"flag"
	"fmt"
	"os"
)

func main() {
	model := flag.String("model", "modelo.json", "Ruta del modelo entrenado (JSON)")
	input := flag.String("input", "aqs_clean.csv", "CSV con filas a predecir (debe incluir la columna target)")
	sample := flag.Int("n", 15, "Filas a mostrar en la tabla predicho vs real")
	maxRows := flag.Int("max-rows", 0, "Maximo de filas a cargar (0 = todas)")
	flag.Parse()

	if err := aqsml.RunPredict(*model, *input, *sample, *maxRows); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
