package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/even-app/even-app/libs/http/middleware"
	"github.com/even-app/even-app/libs/http/server"
	libjwt "github.com/even-app/even-app/libs/jwt"
	"github.com/even-app/even-app/services/api-gateway/internal/config"
	gwmw "github.com/even-app/even-app/services/api-gateway/internal/middleware"
	"github.com/even-app/even-app/services/api-gateway/internal/proxy"
	gwready "github.com/even-app/even-app/services/api-gateway/internal/ready"
	"github.com/even-app/even-app/services/api-gateway/internal/swagger"
)

// New builds the gateway HTTP handler (reverse proxy + JWT + system routes).
func New(cfg config.Config, jwtMgr *libjwt.Manager) (http.Handler, error) {
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
			return nil, err
		}
	}

	merged, err := swagger.FetchMerged(backends)
	if err != nil {
		return nil, err
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

	return middleware.CORS(gwmw.RequireJWT(jwtMgr)(mux)), nil
}
