package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/even-app/even-app/libs/core/logger"
	libjwt "github.com/even-app/even-app/libs/jwt"
	"github.com/even-app/even-app/libs/http/middleware"
	"github.com/even-app/even-app/libs/http/server"
	"github.com/even-app/even-app/libs/postgres"
	"github.com/even-app/even-app/services/lexicon/internal/config"
	lexhandler "github.com/even-app/even-app/services/lexicon/internal/handler"
	http_v1 "github.com/even-app/even-app/services/lexicon/internal/gen/http/v1"
	"github.com/even-app/even-app/services/lexicon/internal/gen/query"
	"github.com/even-app/even-app/services/lexicon/internal/service"
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

	jwtMgr := libjwt.NewManager(cfg.JWTSecret, cfg.AccessTTL())
	ready := func(ctx context.Context) error { return pool.Ping(ctx) }

	querier := query.New(pool)
	lexSvc := service.NewLexiconService(querier)
	httpHandler := lexhandler.NewHTTPHandler(lexSvc)
	secHandler := lexhandler.NewSecurityHandler(jwtMgr)

	oasServer, err := http_v1.NewServer(httpHandler, secHandler)
	if err != nil {
		log.Fatalf("ogen server: %v", err)
	}

	mux := http.NewServeMux()
	server.RegisterHealth(mux, "lexicon", "/api/v1/platform/health")
	server.RegisterReady(mux, ready, "/api/v1/platform/ready")
	mux.Handle("GET /api/v1/openapi.yaml", http_v1.SpecHandler())
	mux.Handle("/", oasServer)

	handler := middleware.CORS(middleware.Recovery(logr, middleware.Logging(logr, mux)))

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
