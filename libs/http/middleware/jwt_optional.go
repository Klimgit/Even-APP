package middleware

import (
	"context"
	"net/http"
	"strings"

	libjwt "github.com/even-app/even-app/libs/jwt"
)

// JWTOptional parses Bearer token when present and stores claims in context.
// Does not reject requests without Authorization (use in handler for protected ops).
func JWTOptional(jwtMgr *libjwt.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if strings.HasPrefix(h, "Bearer ") {
				token := strings.TrimPrefix(h, "Bearer ")
				if claims, err := jwtMgr.ParseAccess(token); err == nil {
					r = r.WithContext(context.WithValue(r.Context(), claimsKey{}, claims))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
