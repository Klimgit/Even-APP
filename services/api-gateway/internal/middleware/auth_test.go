package middleware

import (
	"net/http/httptest"
	"testing"
)

func TestIsPublic(t *testing.T) {
	tests := []struct {
		method string
		path   string
		want   bool
	}{
		{"GET", "/health", true},
		{"GET", "/api/v1/ready", true},
		{"GET", "/api/v1/openapi.yaml", true},
		{"POST", "/api/v1/auth/login", true},
		{"POST", "/api/v1/auth/register", true},
		{"POST", "/api/v1/auth/refresh", true},
		{"GET", "/api/v1/auth/me", false},
		{"GET", "/languages", true},
		{"GET", "/languages/evn", true},
		{"GET", "/languages/evn/alphabet", true},
		{"GET", "/languages/evn/media", false},
		{"GET", "/api/v1/platform/languages/evn/media", false},
		{"GET", "/api/v1/platform/demo/public", true},
		{"GET", "/api/v1/platform/demo/auth", false},
		{"GET", "/api/v1/platform/demo/admin", false},
		{"POST", "/api/v1/platform/media/presign", false},
		{"OPTIONS", "/api/v1/auth/me", false}, // OPTIONS handled in RequireJWT, not IsPublic
	}
	for _, tc := range tests {
		r := httptest.NewRequest(tc.method, tc.path, nil)
		if got := IsPublic(r); got != tc.want {
			t.Errorf("%s %s: got %v want %v", tc.method, tc.path, got, tc.want)
		}
	}
}
