package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/even-app/even-app/libs/http/middleware"
)

// HealthResponse is returned by /health and /api/v1/health.
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// ReadyChecker returns nil when the service is ready to accept traffic.
type ReadyChecker func(ctx context.Context) error

// Options configures the HTTP server.
type Options struct {
	ServiceName      string
	Port             int
	Logger           *slog.Logger
	Ready            ReadyChecker
	OpenAPISpec      []byte
	Handler          http.Handler // if set, used instead of built-in mux
	ExtraHealthPaths []string     // gateway-prefixed aliases, e.g. /api/v1/auth/health
	ExtraReadyPaths  []string
}

// Run starts the HTTP server and blocks until ctx is cancelled, then shuts down gracefully.
func Run(ctx context.Context, opts Options) error {
	var handler http.Handler
	if opts.Handler != nil {
		handler = opts.Handler
	} else {
		mux := http.NewServeMux()
		registerSystemRoutes(mux, opts)
		handler = middleware.CORS(middleware.Recovery(opts.Logger, middleware.Logging(opts.Logger, mux)))
	}

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", opts.Port),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		opts.Logger.Info("listening", "addr", srv.Addr, "service", opts.ServiceName)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		return nil
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

// RegisterHealth mounts GET /health, GET /api/v1/health, and optional extraPaths.
func RegisterHealth(mux *http.ServeMux, serviceName string, extraPaths ...string) {
	writeHealth := func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(HealthResponse{Status: "ok", Service: serviceName})
	}
	paths := append([]string{"/health", "/api/v1/health"}, extraPaths...)
	for _, p := range paths {
		mux.HandleFunc("GET "+p, func(w http.ResponseWriter, _ *http.Request) { writeHealth(w) })
	}
}

// RegisterReady mounts GET /api/v1/ready and optional extraPaths. checker nil means always ready.
func RegisterReady(mux *http.ServeMux, checker ReadyChecker, extraPaths ...string) {
	readyHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if checker != nil {
			if err := checker(r.Context()); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"status": "not_ready",
					"reason": err.Error(),
				})
				return
			}
		}
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	}
	paths := append([]string{"/api/v1/ready"}, extraPaths...)
	for _, p := range paths {
		mux.HandleFunc("GET "+p, readyHandler)
	}
}

func registerSystemRoutes(mux *http.ServeMux, opts Options) {
	RegisterHealth(mux, opts.ServiceName, opts.ExtraHealthPaths...)
	RegisterReady(mux, opts.Ready, opts.ExtraReadyPaths...)

	if len(opts.OpenAPISpec) > 0 {
		spec := opts.OpenAPISpec
		mux.HandleFunc("GET /api/v1/openapi.yaml", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/yaml")
			_, _ = w.Write(spec)
		})
	}
}

// PortString formats listen port.
func PortString(port int) string {
	return strconv.Itoa(port)
}
