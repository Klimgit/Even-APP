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
	"github.com/even-app/even-app/services/auth/internal/config"
	http_v1 "github.com/even-app/even-app/services/auth/internal/gen/http/v1"
	"github.com/even-app/even-app/services/auth/internal/gen/query"
	authhandler "github.com/even-app/even-app/services/auth/internal/handler"
	"github.com/even-app/even-app/services/auth/internal/internalapi"
	"github.com/even-app/even-app/services/auth/internal/repository"
	"github.com/even-app/even-app/services/auth/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
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

	jwtMgr := libjwt.NewManager(cfg.JWTSecret, cfg.AccessTTL)
	ready := func(ctx context.Context) error { return pool.Ping(ctx) }

	querier := query.New(pool)

	var statsSource repository.StatsSource
	if cfg.HasContentHTTP() || cfg.HasLearningHTTP() {
		statsSource = repository.NewPlatformStatsRemote(cfg.ContentServiceURL, cfg.LearningServiceURL, cfg.InternalServiceToken)
	} else {
		var contentPool, learningPool *pgxpool.Pool
		if cfg.ContentDatabaseURL != "" {
			cp, err := postgres.NewPool(ctx, cfg.ContentDatabaseURL)
			if err != nil {
				log.Fatalf("content database: %v", err)
			}
			defer cp.Close()
			contentPool = cp
		}
		if cfg.LearningDatabaseURL != "" {
			lp, err := postgres.NewPool(ctx, cfg.LearningDatabaseURL)
			if err != nil {
				log.Fatalf("learning database: %v", err)
			}
			defer lp.Close()
			learningPool = lp
		}
		statsSource = repository.NewPlatformStatsReader(contentPool, learningPool)
	}

	authSvc := service.NewAuthService(querier, jwtMgr, cfg.RefreshTTL, statsSource)
	httpHandler := authhandler.NewHTTPHandler(authSvc)
	secHandler := authhandler.NewSecurityHandler(jwtMgr)
	internalHandler := internalapi.New(querier)

	oasServer, err := http_v1.NewServer(httpHandler, secHandler)
	if err != nil {
		log.Fatalf("ogen server: %v", err)
	}

	mux := http.NewServeMux()
	server.RegisterHealth(mux, "auth", "/api/v1/auth/health")
	server.RegisterReady(mux, ready, "/api/v1/auth/ready")
	mux.Handle("GET /api/v1/openapi.yaml", http_v1.SpecHandler())
	mux.Handle("/api/v1/internal/", middleware.RequireInternalToken(http.StripPrefix("/api/v1/internal", internalHandler)))
	mux.Handle("/", oasServer)

	handler := middleware.Recovery(logr, middleware.Logging(logr, middleware.AuthRateLimit(mux)))

	if err := server.Run(ctx, server.Options{
		ServiceName: "auth",
		Port:        cfg.Base.HTTPPort,
		Logger:      logr,
		Handler:     handler,
	}); err != nil {
		logr.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
