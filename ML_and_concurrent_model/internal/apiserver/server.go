package apiserver

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// Config holds all runtime parameters for the API server.
type Config struct {
	Port      string
	MongoURL  string
	MongoDB   string
	RedisAddr string
	JWTSecret string
	AdminUser string
	AdminPass string
	NodeAddrs []string // e.g. ["ml-node-1:9000", "ml-node-2:9000"]
	InputPath string   // CSV path accessible inside the container
}

// Server is the main API server.
type Server struct {
	cfg   Config
	hub   *Hub
	store *Store
}

// effectiveNodes returns the dynamic node list from Redis if set,
// otherwise falls back to the static list in cfg.NodeAddrs.
func (s *Server) effectiveNodes(ctx context.Context) []string {
	if nodes, ok := s.store.GetDynamicNodes(ctx); ok && len(nodes) > 0 {
		return nodes
	}
	return s.cfg.NodeAddrs
}

// New creates and initialises a Server, connecting to MongoDB and Redis.
func New(cfg Config) (*Server, error) {
	store, err := NewStore(cfg.MongoURL, cfg.MongoDB, cfg.RedisAddr)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}
	return &Server{cfg: cfg, hub: newHub(), store: store}, nil
}

// Start registers all routes and listens on cfg.Port.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Public
	mux.HandleFunc("POST /api/auth/login", s.handleLogin)
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /ws", s.hub.serveWS)
	mux.HandleFunc("GET /api/districts", s.handleListDistricts)
	mux.HandleFunc("GET /api/districts/predict", s.handleDistrictPredictions)
	mux.HandleFunc("POST /api/predict/public", s.handlePredictPublic)
	mux.HandleFunc("GET /api/stats", s.handleStats)

	// Protected — wrap each group with JWT middleware
	protected := func(method, pattern string, h http.HandlerFunc) {
		mux.Handle(method+" "+pattern, s.jwtMiddleware(h))
	}
	protected("POST", "/api/train", s.handleTrain)
	protected("GET", "/api/train/{jobId}", s.handleTrainStatus)
	protected("GET", "/api/models", s.handleListModels)
	protected("POST", "/api/predict", s.handlePredict)
	protected("GET", "/api/cluster/nodes", s.handleClusterNodes)
	protected("POST", "/api/cluster/nodes", s.handleSetNodes)
	protected("DELETE", "/api/cluster/nodes", s.handleResetNodes)

	handler := corsMiddleware(mux)
	fmt.Printf("API escuchando en %s\n", s.cfg.Port)
	return http.ListenAndServe(s.cfg.Port, handler)
}

// corsMiddleware adds permissive CORS headers for local development.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", strings.Join([]string{
			http.MethodGet, http.MethodPost, http.MethodOptions,
		}, ", "))

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
