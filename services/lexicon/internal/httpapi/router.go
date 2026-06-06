package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/even-app/even-app/libs/http/middleware"
	"github.com/even-app/even-app/libs/http/server"
	libjwt "github.com/even-app/even-app/libs/jwt"
	libs3 "github.com/even-app/even-app/libs/s3"
	"github.com/even-app/even-app/services/lexicon/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewMux(log *slog.Logger, pool *pgxpool.Pool, s3c *libs3.Client, bucket string, userQuotaBytes int64, jwtMgr *libjwt.Manager, openAPI []byte, ready server.ReadyChecker) http.Handler {
	mux := http.NewServeMux()
	server.RegisterHealth(mux, "lexicon")
	server.RegisterReady(mux, ready)
	if len(openAPI) > 0 {
		spec := openAPI
		mux.HandleFunc("GET /api/v1/openapi.yaml", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/yaml")
			_, _ = w.Write(spec)
		})
	}

	jwtMW := middleware.JWT(jwtMgr)
	ms := store.NewMediaStore(pool)
	(&PlatformMediaHandler{Store: ms, S3: s3c, Bucket: bucket, UserQuotaBytes: userQuotaBytes}).Register(mux, jwtMW)

	(&DemoHandler{Store: store.NewDemoStore(pool)}).Register(mux, jwtMW)

	return middleware.CORS(middleware.Recovery(log, middleware.Logging(log, mux)))
}
