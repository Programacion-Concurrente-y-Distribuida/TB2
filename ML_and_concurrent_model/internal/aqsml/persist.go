package aqsml

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gonum.org/v1/gonum/mat"
)

// SavedModel contiene todo lo necesario para reconstruir y aplicar un modelo
// de regresion lineal ya entrenado sin volver a entrenar: los coeficientes
// (beta), los parametros de estandarizacion (medias y desviaciones), el plan
// de codificacion categorica y la configuracion de columnas usada.
type SavedModel struct {
	TargetCol       string           `json:"target_col"`
	Solver          string           `json:"solver"`
	RidgeLambda     float64          `json:"ridge_lambda"`
	NumericFeatures []string         `json:"numeric_features"`
	CatFeatures     []string         `json:"cat_features"`
	OrdFeatures     []string         `json:"ord_features"`
	MaxCatLevels    int              `json:"max_cat_levels"`
	Plan            *catEncodingPlan `json:"plan"`
	Means           []float64        `json:"means"`
	Stds            []float64        `json:"stds"`
	FeatureNames    []string         `json:"feature_names"`
	Beta            []float64        `json:"beta"`
	TrainRows       int              `json:"train_rows"`
	TrainMetrics    metrics          `json:"train_metrics"`
	TestMetrics     metrics          `json:"test_metrics"`
	SavedAt         string           `json:"saved_at"`
}

// saveModel serializa el modelo entrenado a un archivo JSON.
func saveModel(path string, sm *SavedModel) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creando archivo de modelo: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(sm); err != nil {
		return fmt.Errorf("serializando modelo: %w", err)
	}
	return nil
}

// LoadModel lee un modelo entrenado desde un archivo JSON.
func LoadModel(path string) (*SavedModel, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("abriendo modelo: %w", err)
	}
	var sm SavedModel
	if err := json.Unmarshal(data, &sm); err != nil {
		return nil, fmt.Errorf("parseando modelo: %w", err)
	}
	if len(sm.Beta) == 0 {
		return nil, fmt.Errorf("el modelo no contiene coeficientes (beta)")
	}
	return &sm, nil
}

// betaDense reconstruye el vector de coeficientes como matriz columna.
func (sm *SavedModel) betaDense() *mat.Dense {
	return mat.NewDense(len(sm.Beta), 1, append([]float64(nil), sm.Beta...))
}

// RunPredict carga un modelo ya entrenado y lo aplica sobre un CSV de entrada
// SIN volver a entrenar: reconstruye la matriz de diseno con el plan guardado,
// aplica la estandarizacion guardada y calcula las predicciones.
func RunPredict(modelPath, inputPath string, sampleN, maxRows int) error {
	sm, err := LoadModel(modelPath)
	if err != nil {
		return err
	}

	cfg := DefaultConfig()
	cfg.InputPath = inputPath
	cfg.TargetCol = sm.TargetCol
	cfg.NumericFeatures = sm.NumericFeatures
	cfg.CatFeatures = sm.CatFeatures
	cfg.OrdFeatures = sm.OrdFeatures
	cfg.MaxCatLevels = sm.MaxCatLevels
	cfg.MaxRows = maxRows

	rows, loaded, skipped, err := loadCSV(cfg)
	if err != nil {
		return fmt.Errorf("error cargando CSV: %w", err)
	}
	if loaded == 0 {
		return fmt.Errorf("no se cargaron filas validas")
	}

	start := time.Now()
	x := buildDesignMatrix(rows, sm.Plan)
	xStd := standardizeApply(x, sm.Means, sm.Stds)
	xInt := addIntercept(xStd)
	beta := sm.betaDense()
	preds := predict(xInt, beta)
	infer := time.Since(start)

	yTrue := buildTargetVector(rows)
	m := computeMetrics(yTrue, preds)

	fmt.Println("============================================================")
	fmt.Println("Inferencia con modelo YA ENTRENADO (Go + gonum)")
	fmt.Println("============================================================")
	fmt.Printf("Modelo: %s\n", modelPath)
	if sm.SavedAt != "" {
		fmt.Printf("Entrenado: %s | solver=%s | ridge-lambda=%g\n", sm.SavedAt, sm.Solver, sm.RidgeLambda)
	}
	fmt.Printf("Input:  %s\n", inputPath)
	fmt.Printf("Target: %s\n", sm.TargetCol)
	fmt.Printf("Filas cargadas: %d | Filas omitidas: %d\n", loaded, skipped)
	fmt.Printf("Coeficientes del modelo (con intercepto): %d\n", len(sm.Beta))
	fmt.Printf("Metricas del entrenamiento original -> Test R2: %.4f | MAE: %.4f | RMSE: %.4f\n",
		sm.TestMetrics.R2, sm.TestMetrics.MAE, sm.TestMetrics.RMSE)

	r, _ := preds.Dims()
	n := sampleN
	if n > r {
		n = r
	}
	if n > 0 {
		fmt.Printf("\n--- Predicciones (primeras %d filas) ---\n", n)
		fmt.Printf("%4s  %14s  %14s  %14s\n", "#", "Predicho", "Real", "Error")
		for i := 0; i < n; i++ {
			p := preds.At(i, 0)
			t := yTrue.At(i, 0)
			fmt.Printf("%4d  %14.6f  %14.6f  %14.6f\n", i+1, p, t, t-p)
		}
	}

	fmt.Println("\n--- Metricas sobre el input (sin reentrenamiento) ---")
	fmt.Printf("MAE: %.6f | RMSE: %.6f | R2: %.6f\n", m.MAE, m.RMSE, m.R2)
	fmt.Printf("Tiempo de inferencia: %.3fs sobre %d filas\n", infer.Seconds(), r)
	fmt.Println("============================================================")
	return nil
}
