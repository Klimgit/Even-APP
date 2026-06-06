package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/even-app/even-app/libs/core/logger"
	libjwt "github.com/even-app/even-app/libs/jwt"
	"github.com/even-app/even-app/libs/http/server"
	"github.com/even-app/even-app/libs/postgres"
	libs3 "github.com/even-app/even-app/libs/s3"
	apiv1 "github.com/even-app/even-app/services/lexicon/api/http/v1"
	"github.com/even-app/even-app/services/lexicon/internal/config"
	"github.com/even-app/even-app/services/lexicon/internal/httpapi"
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

	s3c, err := libs3.New(cfg.S3)
	if err != nil {
		log.Fatalf("s3: %v", err)
	}

	jwtMgr := libjwt.NewManager(cfg.JWTSecret, cfg.AccessTTL())
	ready := func(ctx context.Context) error { return pool.Ping(ctx) }
	handler := httpapi.NewMux(logr, pool, s3c, cfg.S3.Bucket, cfg.Media.UserQuotaBytes, jwtMgr, apiv1.OpenAPISpec, ready)

	if err := server.Run(ctx, server.Options{
		ServiceName: "lexicon",
		Port:        cfg.Base.HTTPPort,
		Logger:      logr,
		Handler:     handler,
	}); err != nil {
		logr.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
