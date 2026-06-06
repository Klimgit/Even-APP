package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/even-app/even-app/libs/core/logger"
	"github.com/even-app/even-app/libs/http/middleware"
	"github.com/even-app/even-app/libs/http/server"
	libjwt "github.com/even-app/even-app/libs/jwt"
	"github.com/even-app/even-app/services/api-gateway/internal/config"
	gwmw "github.com/even-app/even-app/services/api-gateway/internal/middleware"
	"github.com/even-app/even-app/services/api-gateway/internal/proxy"
	gwready "github.com/even-app/even-app/services/api-gateway/internal/ready"
	"github.com/even-app/even-app/services/api-gateway/internal/swagger"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	logr := logger.New(cfg.Base.LogLevel)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	mux := http.NewServeMux()

	backends := map[string]string{
		"auth":     cfg.AuthURL,
		"lexicon":  cfg.LexiconURL,
		"content":  cfg.ContentURL,
		"learning": cfg.LearningURL,
	}

	routes := []struct {
		prefix string
		target string
	}{
		{"/api/v1/auth/", cfg.AuthURL + "/"},
		{"/api/v1/platform/", cfg.LexiconURL + "/"},
		{"/api/v1/teacher/", cfg.ContentURL + "/"},
		{"/api/v1/courses/", cfg.LearningURL + "/"},
		{"/api/v1/lessons/", cfg.LearningURL + "/"},
		{"/api/v1/progress/", cfg.LearningURL + "/"},
		{"/api/v1/review/", cfg.LearningURL + "/"},
		{"/api/v1/dictionary/", cfg.LearningURL + "/"},
		{"/languages/", cfg.LexiconURL + "/"},
	}
	for _, r := range routes {
		if err := proxy.Mount(mux, r.prefix, r.target); err != nil {
			log.Fatalf("proxy %s: %v", r.prefix, err)
		}
	}

	merged, err := swagger.FetchMerged(backends)
	if err != nil {
		log.Fatalf("swagger: %v", err)
	}

	mux.HandleFunc("GET /api/v1/openapi.yaml", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write(merged)
	})

	mux.HandleFunc("GET /api/v1/gateway/status", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":   "ok",
			"service":  "api-gateway",
			"backends": backends,
		})
	})

	server.RegisterHealth(mux, "api-gateway")
	server.RegisterReady(mux, func(ctx context.Context) error {
		return gwready.CheckBackends(ctx, backends)
	})

	jwtMgr := libjwt.NewManager(cfg.JWTSecret, 0)
	handler := middleware.CORS(middleware.Recovery(logr, middleware.Logging(logr,
		gwmw.RequireJWT(jwtMgr)(mux),
	)))

	if err := server.Run(ctx, server.Options{
		ServiceName: "api-gateway",
		Port:        cfg.Base.HTTPPort,
		Logger:      logr,
		Handler:     handler,
	}); err != nil {
		logr.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
