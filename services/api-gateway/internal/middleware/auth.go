package middleware

import (
	"net/http"
	"strings"

	httpmw "github.com/even-app/even-app/libs/http/middleware"
	libjwt "github.com/even-app/even-app/libs/jwt"
)

// RequireJWT validates Bearer tokens on protected routes. Public routes and OPTIONS pass through.
func RequireJWT(jwtMgr *libjwt.Manager) func(http.Handler) http.Handler {
	jwtMW := httpmw.JWT(jwtMgr)
	return func(next http.Handler) http.Handler {
		protected := jwtMW(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions || IsPublic(r) {
				next.ServeHTTP(w, r)
				return
			}
			protected.ServeHTTP(w, r)
		})
	}
}

// IsPublic reports routes that do not require JWT (see API.md).
func IsPublic(r *http.Request) bool {
	path := r.URL.Path
	switch r.Method {
	case http.MethodGet:
		switch path {
		case "/health", "/api/v1/health", "/api/v1/ready", "/api/v1/openapi.yaml", "/api/v1/gateway/status",
			"/api/v1/auth/demo/public":
			return true
		}
		if isPublicUpstreamSystemGET(path) {
			return true
		}
		return isPublicLanguageGET(path)
	case http.MethodPost:
		switch path {
		case "/api/v1/auth/register", "/api/v1/auth/login", "/api/v1/auth/refresh":
			return true
		}
	}
	return false
}

// isPublicUpstreamSystemGET matches proxied service health/ready, e.g. /api/v1/auth/health.
func isPublicUpstreamSystemGET(path string) bool {
	if !strings.HasPrefix(path, "/api/v1/") {
		return false
	}
	rest := strings.TrimPrefix(path, "/api/v1/")
	if rest == "health" || rest == "ready" {
		return false // gateway's own routes
	}
	return strings.HasSuffix(path, "/health") || strings.HasSuffix(path, "/ready")
}

func isPublicLanguageGET(path string) bool {
	if path == "/languages" {
		return true
	}
	if !strings.HasPrefix(path, "/languages/") {
		return false
	}
	rest := strings.Trim(strings.TrimPrefix(path, "/languages/"), "/")
	if rest == "" {
		return true
	}
	parts := strings.Split(rest, "/")
	switch len(parts) {
	case 1:
		return true // /languages/{code}
	case 2:
		return parts[1] == "alphabet" // /languages/{code}/alphabet
	default:
		return false
	}
}
