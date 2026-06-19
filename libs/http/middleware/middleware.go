package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// Logging wraps a handler with request logging.
func Logging(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r)
		args := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.status,
			"duration_ms", time.Since(start).Milliseconds(),
		}
		if id, ok := RequestIDFromContext(r.Context()); ok {
			args = append(args, "request_id", id)
		}
		log.Info("request", args...)
	})
}

// Recovery catches panics and returns 500.
func Recovery(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error("panic", "err", rec, "stack", string(debug.Stack()))
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

var (
	corsOnce     sync.Once
	corsAllowAll bool
	corsOrigins  map[string]struct{}
)

func loadCORSConfig() {
	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	if strings.TrimSpace(raw) == "" {
		corsAllowAll = true
		return
	}
	corsOrigins = make(map[string]struct{})
	for _, part := range strings.Split(raw, ",") {
		origin := strings.TrimSpace(part)
		if origin != "" {
			corsOrigins[origin] = struct{}{}
		}
	}
}

func resolveCORSOrigin(r *http.Request) string {
	corsOnce.Do(loadCORSConfig)
	if corsAllowAll {
		return "*"
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return ""
	}
	if _, ok := corsOrigins[origin]; ok {
		return origin
	}
	return ""
}

// CORS sets Access-Control headers. When CORS_ALLOWED_ORIGINS is unset, allows * (dev).
// When set (comma-separated), echoes the request Origin only if it matches the allowlist.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := resolveCORSOrigin(r)
		if r.Method == http.MethodOptions {
			setCORSHeaders(w, origin)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(&corsWriter{ResponseWriter: w, origin: origin}, r)
	})
}

func setCORSHeaders(w http.ResponseWriter, origin string) {
	h := w.Header()
	h.Del("Access-Control-Allow-Origin")
	h.Del("Access-Control-Allow-Methods")
	h.Del("Access-Control-Allow-Headers")
	h.Del("Vary")
	if origin == "" {
		return
	}
	h.Set("Access-Control-Allow-Origin", origin)
	if origin != "*" {
		h.Set("Vary", "Origin")
	}
	h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
}

type corsWriter struct {
	http.ResponseWriter
	origin        string
	headerWritten bool
}

func (w *corsWriter) WriteHeader(statusCode int) {
	if !w.headerWritten {
		setCORSHeaders(w.ResponseWriter, w.origin)
		w.headerWritten = true
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *corsWriter) Write(b []byte) (int, error) {
	if !w.headerWritten {
		setCORSHeaders(w.ResponseWriter, w.origin)
		w.headerWritten = true
	}
	return w.ResponseWriter.Write(b)
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
