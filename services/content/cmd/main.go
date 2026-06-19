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
	libjwt "github.com/even-app/even-app/libs/jwt"
	"github.com/even-app/even-app/libs/postgres"
	"github.com/even-app/even-app/services/content/internal/config"
	http_v1 "github.com/even-app/even-app/services/content/internal/gen/http/v1"
	"github.com/even-app/even-app/services/content/internal/gen/query"
	contenthandler "github.com/even-app/even-app/services/content/internal/handler"
	"github.com/even-app/even-app/services/content/internal/internalapi"
	"github.com/even-app/even-app/services/content/internal/repository"
	"github.com/even-app/even-app/services/content/internal/service"
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

	var learningSource repository.LearningSource
	if cfg.HasLearningHTTP() {
		learningSource = repository.NewLearningRemote(cfg.LearningServiceURL, cfg.InternalServiceToken)
	} else if cfg.HasLearningDB() {
		lp, err := postgres.NewPool(ctx, cfg.LearningDatabaseURL)
		if err != nil {
			log.Fatalf("learning database: %v", err)
		}
		defer lp.Close()
		learningSource = repository.NewLearningReader(lp)
	}

	var authSource repository.AuthSource
	if cfg.HasAuthHTTP() {
		authSource = repository.NewAuthRemote(cfg.AuthServiceURL, cfg.InternalServiceToken)
	} else if cfg.HasAuthDB() {
		ap, err := postgres.NewPool(ctx, cfg.AuthDatabaseURL)
		if err != nil {
			log.Fatalf("auth database: %v", err)
		}
		defer ap.Close()
		authSource = repository.NewAuthReader(ap)
	}

	jwtMgr := libjwt.NewManager(cfg.JWTSecret, cfg.AccessTTL())
	ready := func(ctx context.Context) error { return pool.Ping(ctx) }

	querier := query.New(pool)
	contentSvc := service.NewContentService(querier, learningSource, authSource)
	httpHandler := contenthandler.NewHTTPHandler(contentSvc)
	secHandler := contenthandler.NewSecurityHandler(jwtMgr)
	internalHandler := internalapi.New(querier, contentSvc)

	oasServer, err := http_v1.NewServer(httpHandler, secHandler)
	if err != nil {
		log.Fatalf("ogen server: %v", err)
	}

	mux := http.NewServeMux()
	server.RegisterHealth(mux, "content", "/api/v1/teacher/health")
	server.RegisterReady(mux, ready, "/api/v1/teacher/ready")
	mux.Handle("GET /api/v1/openapi.yaml", http_v1.SpecHandler())
	mux.Handle("/api/v1/internal/", middleware.RequireInternalToken(http.StripPrefix("/api/v1/internal", internalHandler)))
	mux.Handle("/", oasServer)

	handler := middleware.Recovery(logr, middleware.Logging(logr, mux))

	if err := server.Run(ctx, server.Options{
		ServiceName: "content",
		Port:        cfg.Base.HTTPPort,
		Logger:      logr,
		Handler:     handler,
	}); err != nil {
		logr.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
