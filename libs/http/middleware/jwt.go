package middleware

import (
	"context"
	"net/http"
	"strings"

	libjwt "github.com/even-app/even-app/libs/jwt"
)

type claimsKey struct{}

// ClaimsFromContext returns JWT claims set by JWT middleware.
func ClaimsFromContext(ctx context.Context) (libjwt.Claims, bool) {
	c, ok := ctx.Value(claimsKey{}).(libjwt.Claims)
	return c, ok
}

// JWT validates Bearer access tokens.
func JWT(jwtMgr *libjwt.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if !strings.HasPrefix(h, "Bearer ") {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(h, "Bearer ")
			claims, err := jwtMgr.ParseAccess(token)
			if err != nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claimsKey{}, claims)))
		})
	}
}
