package apiserver

import (
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"aqsml/internal/aqsml"
	"aqsml/internal/coordinator"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mongoopts "go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ---- response helpers ----

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ---- POST /api/auth/login ----

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	if body.Username != s.cfg.AdminUser || body.Password != s.cfg.AdminPass {
		writeError(w, http.StatusUnauthorized, "credenciales incorrectas")
		return
	}
	token, err := generateToken(body.Username, s.cfg.JWTSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error generando token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

// ---- GET /api/health ----

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	nodes := coordinator.CheckNodes(s.effectiveNodes(r.Context()))
	type nodeSt struct {
		Addr      string `json:"addr"`
		Status    string `json:"status"`
		LatencyMs int64  `json:"latencyMs"`
	}
	ns := make([]nodeSt, len(nodes))
	for i, n := range nodes {
		st := "ok"
		if n.Err != nil {
			st = "unreachable"
		}
		ns[i] = nodeSt{Addr: n.Addr, Status: st, LatencyMs: n.Duration.Milliseconds()}
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "nodes": ns})
}

// ---- POST /api/train ----

type trainRequest struct {
	Solver  string  `json:"solver"`   // ridge | svd | normal
	Lambda  float64 `json:"lambda"`
	MaxRows int     `json:"maxRows"` // 0 = no limit
}

func (s *Server) handleTrain(w http.ResponseWriter, r *http.Request) {
	var req trainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	if req.Solver == "" {
		req.Solver = "ridge"
	}
	if req.Lambda == 0 {
		req.Lambda = 1.0
	}

	jobID := fmt.Sprintf("job-%d", time.Now().UnixMilli())
	now := time.Now()
	jobDoc := &JobDoc{
		JobID:     jobID,
		Status:    "pending",
		Progress:  0,
		StartedAt: now,
	}
	if err := s.store.CreateJob(context.Background(), jobDoc); err != nil {
		writeError(w, http.StatusInternalServerError, "error creando job")
		return
	}

	go s.runTrainingJob(jobID, req)

	writeJSON(w, http.StatusAccepted, map[string]string{"jobId": jobID})
}

func (s *Server) runTrainingJob(jobID string, req trainRequest) {
	ctx := context.Background()
	broadcast := func(v any) { s.hub.Broadcast(v) }
	setJob := func(status string, progress int, extra map[string]any) {
		upd := map[string]any{"status": status, "progress": progress}
		for k, v := range extra {
			upd[k] = v
		}
		s.store.UpdateJob(ctx, jobID, upd) //nolint:errcheck
		broadcast(map[string]any{"type": "jobUpdate", "jobId": jobID, "status": status, "progress": progress})
	}

	setJob("running", 5, nil)

	// 1. Prepare data — leer desde MongoDB (colección measurements)
	broadcast(map[string]any{"type": "phase", "jobId": jobID, "phase": "loadingData"})
	cfg := aqsml.DefaultConfig()
	cfg.Solver = req.Solver
	cfg.RidgeLambda = req.Lambda

	cursor, err := s.store.MeasurementsCursor(ctx, req.MaxRows)
	if err != nil {
		setJob("failed", 0, map[string]any{"error": err.Error()})
		broadcast(map[string]any{"type": "error", "jobId": jobID, "message": "error abriendo cursor de MongoDB: " + err.Error()})
		return
	}
	rows, err := aqsml.LoadRowsFromMongo(ctx, cursor, req.MaxRows)
	if err != nil {
		setJob("failed", 0, map[string]any{"error": err.Error()})
		broadcast(map[string]any{"type": "error", "jobId": jobID, "message": "error cargando datos de MongoDB: " + err.Error()})
		return
	}

	data, err := aqsml.PrepareDataFromRows(rows, cfg)
	if err != nil {
		setJob("failed", 0, map[string]any{"error": err.Error()})
		broadcast(map[string]any{"type": "error", "jobId": jobID, "message": err.Error()})
		return
	}
	setJob("running", 20, nil)

	// 2. Distributed fit
	broadcast(map[string]any{"type": "phase", "jobId": jobID, "phase": "distributing"})

	nodeAddrsActive := s.effectiveNodes(ctx)
	var nodesCompleted int64
	total := len(nodeAddrsActive)
	progress := func(addr string, done bool) {
		if done {
			n := int(atomic.AddInt64(&nodesCompleted, 1))
			pct := 20 + (n*50)/total
			setJob("running", pct, nil)
			broadcast(map[string]any{"type": "nodeDone", "jobId": jobID, "node": addr, "progress": pct})
		} else {
			broadcast(map[string]any{"type": "nodeStart", "jobId": jobID, "node": addr})
		}
	}

	xtxTotal, xtyTotal, nodeResults, err := coordinator.DistributedFit(
		nodeAddrsActive,
		data.TrainX, data.TrainY,
		data.TrainRows, data.Cols,
		progress,
	)
	if err != nil {
		setJob("failed", 0, map[string]any{"error": err.Error()})
		broadcast(map[string]any{"type": "error", "jobId": jobID, "message": err.Error()})
		return
	}
	setJob("running", 75, nil)

	// 3. Solve for β
	broadcast(map[string]any{"type": "phase", "jobId": jobID, "phase": "solving"})
	beta, err := aqsml.SolveNormalEquations(xtxTotal, xtyTotal, data.Cols, req.Lambda)
	if err != nil {
		setJob("failed", 0, map[string]any{"error": err.Error()})
		broadcast(map[string]any{"type": "error", "jobId": jobID, "message": err.Error()})
		return
	}

	// 4. Compute metrics on test set
	testPreds := aqsml.PredictFlat(data.TestX, data.TestRows, data.Cols, beta)
	mae, rmse, r2 := aqsml.ComputeMetricsFlat(data.TestY, testPreds)

	// 5. Save model
	nodeAddrs := make([]string, len(nodeResults))
	for i, nr := range nodeResults {
		nodeAddrs[i] = nr.Addr
	}
	modelDoc := &ModelDoc{
		JobID:           jobID,
		Beta:            beta,
		MAE:             mae,
		RMSE:            rmse,
		R2:              r2,
		TrainedAt:       time.Now(),
		Solver:          req.Solver,
		Lambda:          req.Lambda,
		NodesUsed:       nodeAddrs,
		FeatureNames:    data.FeatureNames,
		NumFeatureNames: data.NumFeatureNames,
		CatFeatureNames: data.CatFeatureNames,
		Means:           data.Means,
		Stds:            data.Stds,
		PlanJSON:        data.PlanJSON,
		Cols:            data.Cols,
		TrainRows:       data.TrainRows,
		TestRows:        data.TestRows,
	}
	modelID, err := s.store.SaveModel(ctx, modelDoc)
	if err != nil {
		setJob("failed", 0, map[string]any{"error": err.Error()})
		return
	}

	now := time.Now()
	setJob("done", 100, map[string]any{"modelId": modelID, "finishedAt": now})
	broadcast(map[string]any{
		"type":    "trainingComplete",
		"jobId":   jobID,
		"modelId": modelID,
		"mae":      mae,
		"rmse":     rmse,
		"r2":       r2,
	})
}

// ---- GET /api/train/{job_id} ----

func (s *Server) handleTrainStatus(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("jobId")
	doc, err := s.store.GetJob(r.Context(), jobID)
	if err == mongo.ErrNoDocuments {
		writeError(w, http.StatusNotFound, "job no encontrado")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

// ---- GET /api/models ----

func (s *Server) handleListModels(w http.ResponseWriter, r *http.Request) {
	docs, err := s.store.ListModels(r.Context(), 20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if docs == nil {
		docs = []ModelDoc{}
	}
	writeJSON(w, http.StatusOK, docs)
}

// ---- GET /api/cluster/nodes ----

func (s *Server) handleClusterNodes(w http.ResponseWriter, r *http.Request) {
	addrs := s.effectiveNodes(r.Context())
	defaultAddrs := make(map[string]bool)
	for _, a := range s.cfg.NodeAddrs {
		defaultAddrs[a] = true
	}
	nodes := coordinator.CheckNodes(addrs)
	type nodeInfo struct {
		Addr      string `json:"addr"`
		Status    string `json:"status"`
		LatencyMs int64  `json:"latencyMs"`
		IsDefault bool   `json:"isDefault"`
	}
	result := make([]nodeInfo, len(nodes))
	for i, n := range nodes {
		st := "ok"
		if n.Err != nil {
			st = "unreachable"
		}
		result[i] = nodeInfo{Addr: n.Addr, Status: st, LatencyMs: n.Duration.Milliseconds(), IsDefault: defaultAddrs[n.Addr]}
	}
	_, usingDynamic := s.store.GetDynamicNodes(r.Context())
	writeJSON(w, http.StatusOK, map[string]any{
		"nodes":        result,
		"usingDynamic": usingDynamic,
		"defaults":     s.cfg.NodeAddrs,
	})
}

// ---- POST /api/cluster/nodes ----

func (s *Server) handleSetNodes(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Nodes []string `json:"nodes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	if len(body.Nodes) == 0 {
		writeError(w, http.StatusBadRequest, "se requiere al menos un nodo")
		return
	}
	if err := s.store.SetDynamicNodes(r.Context(), body.Nodes); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"nodes": body.Nodes, "usingDynamic": true})
}

// ---- DELETE /api/cluster/nodes ----

func (s *Server) handleResetNodes(w http.ResponseWriter, r *http.Request) {
	if err := s.store.ResetDynamicNodes(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"nodes": s.cfg.NodeAddrs, "usingDynamic": false})
}

// ---- POST /api/predict ----

type predictRequest struct {
	ModelID   string             `json:"modelId"` // optional; uses latest if empty
	NumValues map[string]float64 `json:"numValues"`
	CatValues map[string]string  `json:"catValues"`
}

func (s *Server) handlePredict(w http.ResponseWriter, r *http.Request) {
	var req predictRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	// Load model
	var model *ModelDoc
	var err error
	if req.ModelID != "" {
		model, err = s.store.GetModel(r.Context(), req.ModelID)
	} else {
		model, err = s.store.LatestModel(r.Context())
	}
	if err == mongo.ErrNoDocuments {
		writeError(w, http.StatusNotFound, "ningún modelo entrenado disponible")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Build cache key
	cacheKey := predictionCacheKey(model.JobID, req.NumValues, req.CatValues)

	// Check Redis cache
	if cached, ok := s.store.GetCachedPrediction(r.Context(), cacheKey); ok {
		writeJSON(w, http.StatusOK, map[string]any{
			"prediction": cached,
			"cached":     true,
			"modelId":   model.ID.Hex(),
		})
		return
	}

	// Encode input
	x, err := aqsml.EncodeInputJSON(
		model.NumFeatureNames,
		model.PlanJSON,
		req.NumValues,
		req.CatValues,
		model.Means,
		model.Stds,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, "error codificando features: "+err.Error())
		return
	}

	// Predict: ŷ = x · β
	var pred float64
	for i, b := range model.Beta {
		if i < len(x) {
			pred += b * x[i]
		}
	}

	// Cache result
	s.store.CachePrediction(r.Context(), cacheKey, pred)

	writeJSON(w, http.StatusOK, map[string]any{
		"prediction": pred,
		"cached":     false,
		"modelId":   model.ID.Hex(),
	})
}

func predictionCacheKey(modelID string, numVals map[string]float64, catVals map[string]string) string {
	data, _ := json.Marshal(map[string]any{"m": modelID, "n": numVals, "c": catVals})
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// ---- POST /api/predict/public ----
// Public endpoint: users can predict AQI for a district + pollutant without auth.

type publicPredictRequest struct {
	DistrictID string `json:"districtId"`
	Pollutant  string `json:"pollutant"`
}

func (s *Server) handlePredictPublic(w http.ResponseWriter, r *http.Request) {
	var req publicPredictRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	if req.Pollutant == "" {
		req.Pollutant = "PM2.5"
	}

	profile, ok := pollutantProfiles[req.Pollutant]
	if !ok {
		writeError(w, http.StatusBadRequest, "contaminante no soportado: "+req.Pollutant)
		return
	}

	var district *limaDistrict
	for i := range limaDistricts {
		if limaDistricts[i].ID == req.DistrictID {
			district = &limaDistricts[i]
			break
		}
	}
	if district == nil {
		writeError(w, http.StatusBadRequest, "distrito no encontrado: "+req.DistrictID)
		return
	}

	model, err := s.store.LatestModel(r.Context())
	if err == mongo.ErrNoDocuments {
		writeError(w, http.StatusNotFound, "ningún modelo entrenado disponible todavía")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	catValues := map[string]string{
		"pollutant":          req.Pollutant,
		"Parameter Code":     profile.ParameterCode,
		"Sample Duration":    profile.SampleDuration,
		"Event Type":         profile.EventType,
		"Units of Measure":   profile.UnitsOfMeasure,
		"Pollutant Standard": profile.PollutantStandard,
	}
	baseNum := map[string]float64{
		"Latitude":  district.Lat,
		"Longitude": district.Lon,
		"Year":      float64(time.Now().Year()),
	}
	for k, v := range profile.Numeric {
		baseNum[k] = v
	}
	numValues := applyDistrictProfile(baseNum, district.ID)

	x, err := aqsml.EncodeInputJSON(model.NumFeatureNames, model.PlanJSON, numValues, catValues, model.Means, model.Stds)
	if err != nil {
		writeError(w, http.StatusBadRequest, "error codificando features: "+err.Error())
		return
	}

	var pred float64
	for i, b := range model.Beta {
		if i < len(x) {
			pred += b * x[i]
		}
	}
	if pred < 0 {
		pred = 0
	}

	level, color := classify(profile.Breakpoints, pred)

	writeJSON(w, http.StatusOK, map[string]any{
		"district":   district.Name,
		"pollutant":  req.Pollutant,
		"prediction": pred,
		"unit":       profile.Unit,
		"level":      level,
		"color":      color,
		"modelId":    model.ID.Hex(),
	})
}

// ---- GET /api/stats ----
// Returns aggregate social impact statistics based on the latest model predictions.

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	// Lima district populations (approximate, INEI 2022 census)
	districtPop := map[string]int{
		"150101": 276857, "150102": 51952, "150103": 693360, "150104": 33903,
		"150105": 77111, "150106": 356950, "150107": 46213, "150108": 352208,
		"150109": 35299, "150110": 534099, "150111": 188355, "150112": 222498,
		"150113": 71337, "150114": 166912, "150115": 221648, "150116": 50228,
		"150117": 370728, "150118": 220936, "150119": 97551, "150120": 54925,
		"150121": 74523, "150122": 85065, "150123": 140803, "150124": 8161,
		"150125": 387151, "150126": 6621, "150127": 8060, "150128": 163001,
		"150129": 8547, "150130": 116366, "150131": 60735, "150132": 1185637,
		"150133": 433158, "150134": 56095, "150135": 722157, "150136": 134436,
		"150137": 204566, "150138": 1088, "150139": 13036, "150140": 348760,
		"150141": 89283, "150142": 473523, "150143": 407997,
	}

	pollutants := []string{"PM2.5", "O3", "NO2", "CO"}
	type levelStat struct {
		Level      string `json:"level"`
		Color      string `json:"color"`
		Districts  int    `json:"districts"`
		Population int    `json:"population"`
	}

	model, err := s.store.LatestModel(r.Context())
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"totalDistricts":   len(limaDistricts),
			"totalPopulation":  11200000,
			"modelAvailable":   false,
			"byPollutant":      nil,
		})
		return
	}

	type pollutantStats struct {
		Pollutant string      `json:"pollutant"`
		Unit      string      `json:"unit"`
		Levels    []levelStat `json:"levels"`
	}

	byPollutant := make([]pollutantStats, 0, len(pollutants))
	totalPop := 0
	for _, id := range []string{"150101"} {
		totalPop += districtPop[id]
	}
	_ = totalPop

	for _, pollutant := range pollutants {
		profile := pollutantProfiles[pollutant]
		levelMap := make(map[string]*levelStat)

		for _, d := range limaDistricts {
			catValues := map[string]string{
				"pollutant":          pollutant,
				"Parameter Code":     profile.ParameterCode,
				"Sample Duration":    profile.SampleDuration,
				"Event Type":         profile.EventType,
				"Units of Measure":   profile.UnitsOfMeasure,
				"Pollutant Standard": profile.PollutantStandard,
			}
			numValues := map[string]float64{
				"Latitude":  d.Lat,
				"Longitude": d.Lon,
				"Year":      float64(time.Now().Year()),
			}
			for k, v := range profile.Numeric {
				numValues[k] = v
			}
			x, err := aqsml.EncodeInputJSON(model.NumFeatureNames, model.PlanJSON, numValues, catValues, model.Means, model.Stds)
			if err != nil {
				continue
			}
			var pred float64
			for i, b := range model.Beta {
				if i < len(x) {
					pred += b * x[i]
				}
			}
			if pred < 0 {
				pred = 0
			}
			level, color := classify(profile.Breakpoints, pred)
			if _, ok := levelMap[level]; !ok {
				levelMap[level] = &levelStat{Level: level, Color: color}
			}
			levelMap[level].Districts++
			levelMap[level].Population += districtPop[d.ID]
		}

		stats := make([]levelStat, 0, len(levelMap))
		for _, v := range levelMap {
			stats = append(stats, *v)
		}
		byPollutant = append(byPollutant, pollutantStats{
			Pollutant: pollutant,
			Unit:      profile.Unit,
			Levels:    stats,
		})
	}

	totalPopAll := 0
	for _, pop := range districtPop {
		totalPopAll += pop
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"totalDistricts":  len(limaDistricts),
		"totalPopulation": totalPopAll,
		"modelAvailable":  true,
		"modelId":         model.ID.Hex(),
		"trainedAt":       model.TrainedAt,
		"byPollutant":     byPollutant,
	})
}

// ---- GET /api/dataset/info ----

func (s *Server) handleDatasetInfo(w http.ResponseWriter, r *http.Request) {
	datasetPath := s.effectiveDatasetPath(r.Context())
	f, err := os.Open(datasetPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "no se pudo abrir el dataset: "+err.Error())
		return
	}
	defer f.Close()

	stat, _ := f.Stat()

	reader := csv.NewReader(f)
	header, err := reader.Read()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error leyendo cabecera: "+err.Error())
		return
	}

	pollutantCol := -1
	yearCol := -1
	meanCol := -1
	for i, h := range header {
		switch h {
		case "pollutant":
			pollutantCol = i
		case "Year":
			yearCol = i
		case "Arithmetic Mean":
			meanCol = i
		}
	}

	pollutantCounts := make(map[string]int)
	yearMin, yearMax := 9999, 0
	totalRows := 0
	var meanSum float64
	meanCount := 0

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		totalRows++
		if pollutantCol >= 0 && pollutantCol < len(rec) {
			pollutantCounts[rec[pollutantCol]]++
		}
		if yearCol >= 0 && yearCol < len(rec) {
			if y, err := strconv.Atoi(rec[yearCol]); err == nil {
				if y < yearMin { yearMin = y }
				if y > yearMax { yearMax = y }
			}
		}
		if meanCol >= 0 && meanCol < len(rec) {
			if v, err := strconv.ParseFloat(rec[meanCol], 64); err == nil {
				meanSum += v
				meanCount++
			}
		}
	}

	type pollutantCount struct {
		Name  string `json:"name"`
		Rows  int    `json:"rows"`
	}
	pollutants := make([]pollutantCount, 0, len(pollutantCounts))
	for k, v := range pollutantCounts {
		pollutants = append(pollutants, pollutantCount{Name: k, Rows: v})
	}

	var globalMean *float64
	if meanCount > 0 {
		m := meanSum / float64(meanCount)
		globalMean = &m
	}

	fileSizeMB := float64(0)
	if stat != nil {
		fileSizeMB = float64(stat.Size()) / 1024 / 1024
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"path":        datasetPath,
		"fileSizeMB":  fmt.Sprintf("%.1f", fileSizeMB),
		"totalRows":   totalRows,
		"columns":     len(header),
		"columnNames": header,
		"yearMin":     yearMin,
		"yearMax":     yearMax,
		"pollutants":  pollutants,
		"globalMean":  globalMean,
	})
}

// ---- GET /api/dataset/list (protected) ----

func (s *Server) handleDatasetList(w http.ResponseWriter, r *http.Request) {
	dataDir := filepath.Dir(s.cfg.InputPath)
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "no se pudo leer el directorio de datos: "+err.Error())
		return
	}

	type datasetEntry struct {
		Name       string `json:"name"`
		Path       string `json:"path"`
		SizeMB     string `json:"sizeMB"`
		IsSelected bool   `json:"isSelected"`
	}

	selected := s.effectiveDatasetPath(r.Context())
	datasets := make([]datasetEntry, 0)
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".csv" {
			continue
		}
		fullPath := filepath.Join(dataDir, e.Name())
		info, err := e.Info()
		sizeMB := "?"
		if err == nil {
			sizeMB = fmt.Sprintf("%.1f", float64(info.Size())/1024/1024)
		}
		datasets = append(datasets, datasetEntry{
			Name:       e.Name(),
			Path:       fullPath,
			SizeMB:     sizeMB,
			IsSelected: fullPath == selected,
		})
	}

	writeJSON(w, http.StatusOK, datasets)
}

// ---- POST /api/dataset/select (protected) ----

func (s *Server) handleDatasetSelect(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	if req.Path == "" {
		writeError(w, http.StatusBadRequest, "path es requerido")
		return
	}
	if _, err := os.Stat(req.Path); err != nil {
		writeError(w, http.StatusBadRequest, "archivo no encontrado: "+req.Path)
		return
	}
	if err := s.store.SetSelectedDataset(r.Context(), req.Path); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"selected": req.Path})
}

// ---- GET /api/dataset/status (protected) ----

func (s *Server) handleDatasetStatus(w http.ResponseWriter, r *http.Request) {
	count, err := s.store.MeasurementsCount(r.Context())
	if err != nil {
		count = 0
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"measurementsCount": count,
		"ready":             count > 0,
	})
}

// ---- POST /api/dataset/migrate (protected) ----
// Runs CSV → MongoDB migration in the background; streams progress via WebSocket.

const migrateBatchSize = 2000

type columnKind int

const (
	kindString columnKind = iota
	kindInt
	kindFloat
	kindBool
)

var columnKinds = map[string]columnKind{
	"State Code": kindString, "County Code": kindString, "Site Num": kindString,
	"POC": kindInt, "Parameter Code": kindInt, "Year": kindInt,
	"Latitude": kindFloat, "Longitude": kindFloat,
	"Parameter Name": kindString, "Sample Duration": kindString,
	"Pollutant Standard": kindString, "Units of Measure": kindString,
	"Event Type": kindString, "State Name": kindString, "County Name": kindString,
	"City Name": kindString, "CBSA Name": kindString,
	"pollutant": kindString, "is_synthetic": kindBool,
}

func kindFor(col string) columnKind {
	if k, ok := columnKinds[col]; ok {
		return k
	}
	return kindFloat
}

func convertCell(kind columnKind, raw string) any {
	s := strings.TrimSpace(raw)
	missing := s == "" || strings.EqualFold(s, "nan") || strings.EqualFold(s, "null") || strings.EqualFold(s, "none")
	switch kind {
	case kindString:
		return raw
	case kindBool:
		switch strings.ToLower(s) {
		case "true", "1":
			return true
		case "false", "0":
			return false
		}
		return nil
	case kindInt:
		if missing {
			return nil
		}
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			return n
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil && !math.IsNaN(f) {
			return int64(f)
		}
		return nil
	case kindFloat:
		if missing {
			return nil
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil && !math.IsNaN(f) && !math.IsInf(f, 0) {
			return f
		}
		return nil
	}
	return raw
}

func (s *Server) handleDatasetMigrate(w http.ResponseWriter, r *http.Request) {
	path := s.effectiveDatasetPath(r.Context())
	if _, err := os.Stat(path); err != nil {
		writeError(w, http.StatusBadRequest, "dataset no encontrado: "+path)
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{"status": "migrating", "path": path})

	go s.runMigration(path)
}

func (s *Server) runMigration(csvPath string) {
	ctx := context.Background()
	broadcast := func(v any) { s.hub.Broadcast(v) }

	broadcast(map[string]any{"type": "migrateStart", "path": csvPath})

	// Drop existing collection
	if err := s.store.DropMeasurements(ctx); err != nil {
		broadcast(map[string]any{"type": "migrateError", "message": "error al limpiar colección: " + err.Error()})
		return
	}

	f, err := os.Open(csvPath)
	if err != nil {
		broadcast(map[string]any{"type": "migrateError", "message": "error abriendo CSV: " + err.Error()})
		return
	}
	defer f.Close()

	stat, _ := f.Stat()
	totalBytes := int64(0)
	if stat != nil {
		totalBytes = stat.Size()
	}

	reader := csv.NewReader(f)
	reader.ReuseRecord = true

	header, err := reader.Read()
	if err != nil {
		broadcast(map[string]any{"type": "migrateError", "message": "error leyendo cabecera: " + err.Error()})
		return
	}
	cols := make([]string, len(header))
	copy(cols, header)
	kinds := make([]columnKind, len(cols))
	for i, c := range cols {
		kinds[i] = kindFor(c)
	}

	coll := s.store.col("measurements")
	insertOpts := mongoopts.InsertMany().SetOrdered(false)

	var inserted atomic.Int64
	var skipped atomic.Int64
	batches := make(chan []any, 4)
	var wg sync.WaitGroup

	// 4 workers de inserción
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range batches {
				res, err := coll.InsertMany(ctx, batch, insertOpts)
				if res != nil {
					inserted.Add(int64(len(res.InsertedIDs)))
				}
				if err != nil {
					var bwe mongo.BulkWriteException
					if errors.As(err, &bwe) {
						skipped.Add(int64(len(bwe.WriteErrors)))
					}
				}
			}
		}()
	}

	// Reporte de progreso cada 2s
	ticker := time.NewTicker(2 * time.Second)
	done := make(chan struct{})
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				n := inserted.Load()
				pct := 0
				if totalBytes > 0 {
					// estimación aproximada basada en docs insertados vs total esperado
					pct = int(math.Min(float64(n)/3000000*100, 99))
				}
				broadcast(map[string]any{
					"type":     "migrateProgress",
					"inserted": n,
					"progress": pct,
				})
			}
		}
	}()

	// Productor
	batch := make([]any, 0, migrateBatchSize)
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			skipped.Add(1)
			continue
		}
		doc := make(bson.D, 0, len(cols))
		for i, col := range cols {
			val := ""
			if i < len(rec) {
				val = rec[i]
			}
			doc = append(doc, bson.E{Key: col, Value: convertCell(kinds[i], val)})
		}
		batch = append(batch, doc)
		if len(batch) >= migrateBatchSize {
			batches <- batch
			batch = make([]any, 0, migrateBatchSize)
		}
	}
	if len(batch) > 0 {
		batches <- batch
	}
	close(batches)
	wg.Wait()
	close(done)

	broadcast(map[string]any{
		"type":     "migrateDone",
		"inserted": inserted.Load(),
		"skipped":  skipped.Load(),
	})
}
