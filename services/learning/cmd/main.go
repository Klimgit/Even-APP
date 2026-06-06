package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/even-app/even-app/libs/core/logger"
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

	if err := server.Run(ctx, server.Options{
		ServiceName: "learning",
		Port:        cfg.Base.HTTPPort,
		Logger:      logr,
		Ready:       ready,
		OpenAPISpec: apiv1.OpenAPISpec,
	}); err != nil {
		logr.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
