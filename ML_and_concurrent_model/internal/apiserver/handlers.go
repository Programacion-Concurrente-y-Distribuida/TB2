package apiserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"aqsml/internal/aqsml"
	"aqsml/internal/coordinator"
	"go.mongodb.org/mongo-driver/v2/mongo"
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
	nodes := coordinator.CheckNodes(s.cfg.NodeAddrs)
	type nodeSt struct {
		Addr      string `json:"addr"`
		Status    string `json:"status"`
		LatencyMs int64  `json:"latency_ms"`
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
	MaxRows int     `json:"max_rows"` // 0 = no limit
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

	writeJSON(w, http.StatusAccepted, map[string]string{"job_id": jobID})
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
		broadcast(map[string]any{"type": "job_update", "job_id": jobID, "status": status, "progress": progress})
	}

	setJob("running", 5, nil)

	// 1. Prepare data
	broadcast(map[string]any{"type": "phase", "job_id": jobID, "phase": "loading_data"})
	cfg := aqsml.DefaultConfig()
	cfg.InputPath = s.cfg.InputPath
	cfg.MaxRows = req.MaxRows
	cfg.Solver = req.Solver
	cfg.RidgeLambda = req.Lambda

	data, err := aqsml.PrepareData(cfg)
	if err != nil {
		setJob("failed", 0, map[string]any{"error": err.Error()})
		broadcast(map[string]any{"type": "error", "job_id": jobID, "message": err.Error()})
		return
	}
	setJob("running", 20, nil)

	// 2. Distributed fit
	broadcast(map[string]any{"type": "phase", "job_id": jobID, "phase": "distributing"})

	var nodesCompleted int64
	total := len(s.cfg.NodeAddrs)
	progress := func(addr string, done bool) {
		if done {
			n := int(atomic.AddInt64(&nodesCompleted, 1))
			pct := 20 + (n*50)/total
			setJob("running", pct, nil)
			broadcast(map[string]any{"type": "node_done", "job_id": jobID, "node": addr, "progress": pct})
		} else {
			broadcast(map[string]any{"type": "node_start", "job_id": jobID, "node": addr})
		}
	}

	xtxTotal, xtyTotal, nodeResults, err := coordinator.DistributedFit(
		s.cfg.NodeAddrs,
		data.TrainX, data.TrainY,
		data.TrainRows, data.Cols,
		progress,
	)
	if err != nil {
		setJob("failed", 0, map[string]any{"error": err.Error()})
		broadcast(map[string]any{"type": "error", "job_id": jobID, "message": err.Error()})
		return
	}
	setJob("running", 75, nil)

	// 3. Solve for β
	broadcast(map[string]any{"type": "phase", "job_id": jobID, "phase": "solving"})
	beta, err := aqsml.SolveNormalEquations(xtxTotal, xtyTotal, data.Cols, req.Lambda)
	if err != nil {
		setJob("failed", 0, map[string]any{"error": err.Error()})
		broadcast(map[string]any{"type": "error", "job_id": jobID, "message": err.Error()})
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
	setJob("done", 100, map[string]any{"model_id": modelID, "finished_at": now})
	broadcast(map[string]any{
		"type":     "training_complete",
		"job_id":   jobID,
		"model_id": modelID,
		"mae":      mae,
		"rmse":     rmse,
		"r2":       r2,
	})
}

// ---- GET /api/train/{job_id} ----

func (s *Server) handleTrainStatus(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("job_id")
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
	writeJSON(w, http.StatusOK, docs)
}

// ---- GET /api/cluster/nodes ----

func (s *Server) handleClusterNodes(w http.ResponseWriter, r *http.Request) {
	nodes := coordinator.CheckNodes(s.cfg.NodeAddrs)
	type nodeInfo struct {
		Addr      string `json:"addr"`
		Status    string `json:"status"`
		LatencyMs int64  `json:"latency_ms"`
	}
	result := make([]nodeInfo, len(nodes))
	for i, n := range nodes {
		st := "ok"
		if n.Err != nil {
			st = "unreachable"
		}
		result[i] = nodeInfo{Addr: n.Addr, Status: st, LatencyMs: n.Duration.Milliseconds()}
	}
	writeJSON(w, http.StatusOK, result)
}

// ---- POST /api/predict ----

type predictRequest struct {
	ModelID   string             `json:"model_id"` // optional; uses latest if empty
	NumValues map[string]float64 `json:"num_values"`
	CatValues map[string]string  `json:"cat_values"`
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
			"model_id":   model.ID.Hex(),
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
		"model_id":   model.ID.Hex(),
	})
}

func predictionCacheKey(modelID string, numVals map[string]float64, catVals map[string]string) string {
	data, _ := json.Marshal(map[string]any{"m": modelID, "n": numVals, "c": catVals})
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
