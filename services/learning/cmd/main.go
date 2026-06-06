package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/even-app/even-app/libs/core/logger"
	"github.com/even-app/even-app/libs/http/middleware"
	"github.com/even-app/even-app/libs/http/server"
	"github.com/even-app/even-app/libs/postgres"
	apiv1 "github.com/even-app/even-app/services/learning/api/http/v1"
	"github.com/even-app/even-app/services/learning/internal/config"
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

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	ready := func(ctx context.Context) error { return pool.Ping(ctx) }

	mux := http.NewServeMux()
	server.RegisterHealth(mux, "learning", "/api/v1/courses/health")
	server.RegisterReady(mux, ready, "/api/v1/courses/ready")
	if len(apiv1.OpenAPISpec) > 0 {
		spec := apiv1.OpenAPISpec
		mux.HandleFunc("GET /api/v1/openapi.yaml", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/yaml")
			_, _ = w.Write(spec)
		})
	}

	handler := middleware.CORS(middleware.Recovery(logr, middleware.Logging(logr, mux)))

	if err := server.Run(ctx, server.Options{
		ServiceName: "learning",
		Port:        cfg.Base.HTTPPort,
		Logger:      logr,
		Handler:     handler,
	}); err != nil {
		logr.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
