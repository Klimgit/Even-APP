package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const authRateLimitPerMin = 30

type tokenBucket struct {
	tokens   float64
	lastFill time.Time
}

type ipRateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket
	rate    float64
	burst   float64
}

func newIPRateLimiter(perMinute int) *ipRateLimiter {
	return &ipRateLimiter{
		buckets: make(map[string]*tokenBucket),
		rate:    float64(perMinute) / 60.0,
		burst:   float64(perMinute),
	}
}

func (l *ipRateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, ok := l.buckets[key]
	if !ok {
		l.buckets[key] = &tokenBucket{tokens: l.burst - 1, lastFill: now}
		return true
	}

	elapsed := now.Sub(b.lastFill).Seconds()
	b.tokens += elapsed * l.rate
	if b.tokens > l.burst {
		b.tokens = l.burst
	}
	b.lastFill = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func isAuthRateLimitedRoute(method, path string) bool {
	if method != http.MethodPost {
		return false
	}
	switch path {
	case "/api/v1/auth/login", "/api/v1/auth/register":
		return true
	default:
		return false
	}
}

// AuthRateLimit applies an in-memory per-IP token bucket to login and register.
func AuthRateLimit(next http.Handler) http.Handler {
	limiter := newIPRateLimiter(authRateLimitPerMin)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isAuthRateLimitedRoute(r.Method, r.URL.Path) {
			if !limiter.allow(clientIP(r)) {
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
