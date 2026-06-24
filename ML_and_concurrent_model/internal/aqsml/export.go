package aqsml

import (
	"encoding/json"
	"fmt"
	"math"

	"gonum.org/v1/gonum/mat"
)

// PreparedData holds preprocessed train/test matrices ready for distributed fitting.
// All matrices are flat row-major float64 slices; Cols includes the intercept column.
type PreparedData struct {
	TrainX          []float64
	TrainY          []float64
	TrainRows       int
	TestX           []float64
	TestY           []float64
	TestRows        int
	Cols            int      // includes intercept (col 0 == 1.0)
	FeatureNames    []string // length Cols-1, excludes intercept
	Means           []float64
	Stds            []float64
	PlanJSON        []byte   // catEncodingPlan as JSON — used by EncodeInputJSON for predictions
	NumFeatureNames []string
	CatFeatureNames []string
}

// PrepareData loads the CSV, encodes features, standardizes, and returns flat float64
// matrices with an intercept column prepended. Safe to call from any goroutine.
func PrepareData(cfg Config) (*PreparedData, error) {
	if len(cfg.NumericFeatures) == 0 && len(cfg.CatFeatures) == 0 {
		return nil, fmt.Errorf("al menos una feature requerida")
	}

	rows, _, _, err := loadCSV(cfg)
	if err != nil {
		return nil, fmt.Errorf("loadCSV: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("sin filas validas")
	}

	trainRows, testRows, err := splitRows(rows, cfg.TrainYearEnd)
	if err != nil {
		return nil, err
	}

	plan, err := buildCatEncodingPlan(trainRows, cfg.CatFeatures, cfg.OrdFeatures, cfg.MaxCatLevels)
	if err != nil {
		return nil, err
	}
	plan.FeatureNames = append(append([]string(nil), cfg.NumericFeatures...), plan.FeatureNames...)

	trainXMat := buildDesignMatrix(trainRows, plan)
	trainYMat := buildTargetVector(trainRows)
	testXMat := buildDesignMatrix(testRows, plan)
	testYMat := buildTargetVector(testRows)

	trainXStd, testXStd, means, stds := standardize(trainXMat, testXMat)

	featureNames := append([]string(nil), plan.FeatureNames...)
	if cfg.UsePCA {
		var k int
		trainXStd, testXStd, k, _, err = applyPCA(trainXStd, testXStd, cfg.PCAVarianceGoal)
		if err != nil {
			return nil, err
		}
		featureNames = make([]string, k)
		for i := range featureNames {
			featureNames[i] = fmt.Sprintf("PC%d", i+1)
		}
	}

	trainXInt := addIntercept(trainXStd)
	testXInt := addIntercept(testXStd)

	tRows, tCols := trainXInt.Dims()
	teRows, _ := testXInt.Dims()

	planJSON, err := json.Marshal(plan)
	if err != nil {
		return nil, fmt.Errorf("serializando plan: %w", err)
	}

	return &PreparedData{
		TrainX:          denseToFlat(trainXInt, tRows, tCols),
		TrainY:          vecToFlat(trainYMat, tRows),
		TrainRows:       tRows,
		TestX:           denseToFlat(testXInt, teRows, tCols),
		TestY:           vecToFlat(testYMat, teRows),
		TestRows:        teRows,
		Cols:            tCols,
		FeatureNames:    featureNames,
		Means:           means,
		Stds:            stds,
		PlanJSON:        planJSON,
		NumFeatureNames: cfg.NumericFeatures,
		CatFeatureNames: cfg.CatFeatures,
	}, nil
}

// SolveNormalEquations resolves β from aggregated XTX (cols×cols) and XTy (cols).
// Ridge regularization is applied here — worker nodes compute raw partials only.
// xtxData is modified in place when lambda > 0.
func SolveNormalEquations(xtxData, xtyData []float64, cols int, lambda float64) ([]float64, error) {
	if lambda > 0 {
		for j := 1; j < cols; j++ { // skip intercept column 0
			xtxData[j*cols+j] += lambda
		}
	}

	xtx := mat.NewSymDense(cols, xtxData)
	xty := mat.NewDense(cols, 1, xtyData)
	beta := mat.NewDense(cols, 1, nil)

	var chol mat.Cholesky
	if ok := chol.Factorize(xtx); ok {
		if err := chol.SolveTo(beta, xty); err == nil {
			return denseColToSlice(beta, cols), nil
		}
	}
	return nil, fmt.Errorf("Cholesky falló; usa ridge con lambda > 0")
}

// ComputeMetricsFlat returns MAE, RMSE and R² for flat prediction/truth slices.
func ComputeMetricsFlat(yTrue, yPred []float64) (mae, rmse, r2 float64) {
	n := len(yTrue)
	if n == 0 {
		return
	}
	var sumAbs, sumSq, sumTrue float64
	for i, t := range yTrue {
		sumTrue += t
		d := t - yPred[i]
		sumAbs += math.Abs(d)
		sumSq += d * d
	}
	mean := sumTrue / float64(n)
	var ssTot float64
	for _, t := range yTrue {
		d := t - mean
		ssTot += d * d
	}
	r2 = 1.0
	if ssTot > 0 {
		r2 = 1 - sumSq/ssTot
	}
	return sumAbs / float64(n), math.Sqrt(sumSq / float64(n)), r2
}

// PredictFlat computes ŷ = X·β for a flat row-major matrix X (rows×cols) and beta (cols).
func PredictFlat(X []float64, rows, cols int, beta []float64) []float64 {
	preds := make([]float64, rows)
	for i := 0; i < rows; i++ {
		var sum float64
		for j := 0; j < cols; j++ {
			sum += X[i*cols+j] * beta[j]
		}
		preds[i] = sum
	}
	return preds
}

// EncodeInputJSON encodes a single prediction input into a feature vector (with intercept)
// using the stored encoding plan and the standardization params from training.
func EncodeInputJSON(
	numFeatNames []string,
	planJSON []byte,
	numValues map[string]float64,
	catValues map[string]string,
	means, stds []float64,
) ([]float64, error) {
	var plan catEncodingPlan
	if err := json.Unmarshal(planJSON, &plan); err != nil {
		return nil, fmt.Errorf("plan JSON inválido: %w", err)
	}

	catCols := 0
	for _, enc := range plan.Nominal {
		catCols += len(enc.ActiveLvls)
	}
	catCols += len(plan.Ordinal)

	vec := make([]float64, len(numFeatNames)+catCols)

	for i, name := range numFeatNames {
		vec[i] = numValues[name]
	}

	offset := len(numFeatNames)
	for _, enc := range plan.Nominal {
		val := catValues[enc.ColName]
		if pos, ok := enc.LvlToOffset[val]; ok {
			vec[offset+pos] = 1.0
		}
		offset += len(enc.ActiveLvls)
	}
	for _, enc := range plan.Ordinal {
		val := catValues[enc.ColName]
		if rank, ok := enc.LevelRank[val]; ok {
			vec[offset] = rank
		}
		offset++
	}

	// Standardize using training statistics
	for j := range vec {
		if j >= len(means) {
			break
		}
		s := stds[j]
		if s == 0 {
			s = 1.0
		}
		vec[j] = (vec[j] - means[j]) / s
	}

	// Prepend intercept
	out := make([]float64, 1+len(vec))
	out[0] = 1.0
	copy(out[1:], vec)
	return out, nil
}

// --- package-private helpers ---

func denseToFlat(m *mat.Dense, rows, cols int) []float64 {
	out := make([]float64, rows*cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			out[i*cols+j] = m.At(i, j)
		}
	}
	return out
}

func vecToFlat(m *mat.Dense, rows int) []float64 {
	out := make([]float64, rows)
	for i := 0; i < rows; i++ {
		out[i] = m.At(i, 0)
	}
	return out
}

func denseColToSlice(m *mat.Dense, rows int) []float64 {
	out := make([]float64, rows)
	for i := 0; i < rows; i++ {
		out[i] = m.At(i, 0)
	}
	return out
}
