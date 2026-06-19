package middleware

import (
	"net/http"
	"os"
)

const internalTokenHeader = "X-Internal-Token"

// RequireInternalToken protects service-to-service routes.
func RequireInternalToken(next http.Handler) http.Handler {
	expected := os.Getenv("INTERNAL_SERVICE_TOKEN")
	if expected == "" {
		expected = "dev-internal-token"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(internalTokenHeader) != expected {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
