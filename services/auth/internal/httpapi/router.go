package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	libjwt "github.com/even-app/even-app/libs/jwt"
	"github.com/even-app/even-app/libs/http/middleware"
	"github.com/even-app/even-app/libs/http/server"
	"github.com/even-app/even-app/services/auth/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewMux(log *slog.Logger, pool *pgxpool.Pool, jwtMgr *libjwt.Manager, refreshTTL time.Duration, openAPI []byte, ready server.ReadyChecker) http.Handler {
	mux := http.NewServeMux()
	server.RegisterHealth(mux, "auth", "/api/v1/auth/health")
	server.RegisterReady(mux, ready, "/api/v1/auth/ready")
	if len(openAPI) > 0 {
		spec := openAPI
		mux.HandleFunc("GET /api/v1/openapi.yaml", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/yaml")
			_, _ = w.Write(spec)
		})
	}

	users := store.NewUserStore(pool)
	jwtMW := middleware.JWT(jwtMgr)
	(&AuthHandler{Users: users, JWT: jwtMgr, RefreshTTL: refreshTTL}).Register(mux, jwtMW)

	return middleware.CORS(middleware.Recovery(log, middleware.Logging(log, mux)))
}
