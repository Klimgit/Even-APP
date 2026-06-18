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
	"github.com/even-app/even-app/services/learning/internal/config"
	http_v1 "github.com/even-app/even-app/services/learning/internal/gen/http/v1"
	learnhandler "github.com/even-app/even-app/services/learning/internal/handler"
	"github.com/even-app/even-app/services/learning/internal/repository"
	"github.com/even-app/even-app/services/learning/internal/service"
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

	var contentPool *repository.ContentReader
	if cfg.HasContentDB() {
		cp, err := postgres.NewPool(ctx, cfg.ContentDatabaseURL)
		if err != nil {
			log.Fatalf("content database: %v", err)
		}
		defer cp.Close()
		contentPool = repository.NewContentReader(cp)
	}

	var lexiconPool *repository.LexiconReader
	if cfg.HasLexiconDB() {
		lp, err := postgres.NewPool(ctx, cfg.LexiconDatabaseURL)
		if err != nil {
			log.Fatalf("lexicon database: %v", err)
		}
		defer lp.Close()
		lexiconPool = repository.NewLexiconReader(lp)
	}

	jwtMgr := libjwt.NewManager(cfg.JWTSecret, cfg.AccessTTL())
	ready := func(ctx context.Context) error { return pool.Ping(ctx) }

	learnSvc := service.NewLearningService(pool, contentPool, lexiconPool)
	httpHandler := learnhandler.NewHTTPHandler(learnSvc)
	secHandler := learnhandler.NewSecurityHandler(jwtMgr)

	oasServer, err := http_v1.NewServer(httpHandler, secHandler)
	if err != nil {
		log.Fatalf("ogen server: %v", err)
	}

	mux := http.NewServeMux()
	server.RegisterHealth(mux, "learning", "/api/v1/courses/health")
	server.RegisterReady(mux, ready, "/api/v1/courses/ready")
	mux.Handle("GET /api/v1/openapi.yaml", http_v1.SpecHandler())
	mux.Handle("/", oasServer)

	handler := middleware.Recovery(logr, middleware.Logging(logr, mux))

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
