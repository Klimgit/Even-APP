package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/even-app/even-app/libs/core/logger"
	"github.com/even-app/even-app/libs/http/middleware"
	"github.com/even-app/even-app/libs/http/server"
	libjwt "github.com/even-app/even-app/libs/jwt"
	"github.com/even-app/even-app/services/api-gateway/internal/config"
	"github.com/even-app/even-app/services/api-gateway/internal/handler"
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

	jwtMgr := libjwt.NewManager(cfg.JWTSecret, 0)
	h, err := handler.New(cfg, jwtMgr)
	if err != nil {
		log.Fatalf("gateway handler: %v", err)
	}
	handlerChain := middleware.Recovery(logr, middleware.Logging(logr, h))

	if err := server.Run(ctx, server.Options{
		ServiceName: "api-gateway",
		Port:        cfg.Base.HTTPPort,
		Logger:      logr,
		Handler:     handlerChain,
	}); err != nil {
		logr.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
