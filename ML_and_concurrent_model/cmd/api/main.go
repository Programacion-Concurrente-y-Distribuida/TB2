// api: HTTP + WebSocket server that coordinates the distributed ML cluster.
//
// Environment variables:
//
//	PORT         listen address          (default :8080)
//	MONGO_URL    MongoDB connection URI  (default mongodb://mongo:27017)
//	MONGO_DB     database name           (default aqs)
//	REDIS_ADDR   Redis address           (default redis:6379; empty = disabled)
//	JWT_SECRET   HMAC-SHA256 secret      (required)
//	ADMIN_USER   login username          (default admin)
//	ADMIN_PASS   login password          (required)
//	NODE_ADDRS   comma-separated list of ml-node addresses (default ml-node-1:9000)
//	INPUT_PATH   CSV file path           (default aqs_final_3M.csv)
package main

import (
	"log"
	"os"
	"strings"

	"aqsml/internal/apiserver"
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	nodeAddrsRaw := env("NODE_ADDRS", "ml-node-1:9000")
	nodeAddrs := strings.Split(nodeAddrsRaw, ",")
	for i := range nodeAddrs {
		nodeAddrs[i] = strings.TrimSpace(nodeAddrs[i])
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET es requerido")
	}
	adminPass := os.Getenv("ADMIN_PASS")
	if adminPass == "" {
		log.Fatal("ADMIN_PASS es requerido")
	}

	cfg := apiserver.Config{
		Port:      env("PORT", ":8080"),
		MongoURL:  env("MONGO_URL", "mongodb://mongo:27017"),
		MongoDB:   env("MONGO_DB", "aqs"),
		RedisAddr: env("REDIS_ADDR", "redis:6379"),
		JWTSecret: jwtSecret,
		AdminUser: env("ADMIN_USER", "admin"),
		AdminPass: adminPass,
		NodeAddrs: nodeAddrs,
		InputPath: env("INPUT_PATH", "aqs_final_3M.csv"),
	}

	srv, err := apiserver.New(cfg)
	if err != nil {
		log.Fatalf("init server: %v", err)
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
