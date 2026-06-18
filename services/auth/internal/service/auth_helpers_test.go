package service

import (
	"errors"
	"strings"
	"testing"
)

func TestIsUniqueViolation(t *testing.T) {
	tests := []struct {
		msg  string
		want bool
	}{
		{"duplicate key value violates unique constraint", true},
		{"ERROR: unique constraint failed", true},
		{"connection refused", false},
		{"", false},
	}
	for _, tc := range tests {
		if got := isUniqueViolation(errors.New(tc.msg)); got != tc.want {
			t.Errorf("isUniqueViolation(%q) = %v, want %v", tc.msg, got, tc.want)
		}
	}
}

func TestHashToken_deterministic(t *testing.T) {
	a := hashToken("refresh-token-abc")
	b := hashToken("refresh-token-abc")
	if a != b || len(a) != 64 {
		t.Fatalf("hashToken not stable hex sha256: %q", a)
	}
	if hashToken("other") == a {
		t.Fatal("different tokens should hash differently")
	}
}

func TestNewRefreshToken_length(t *testing.T) {
	tok, err := newRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(tok) != 64 {
		t.Fatalf("refresh token hex length = %d, want 64", len(tok))
	}
	if strings.ContainsAny(tok, "ghijklmnopqrstuvwxyz") {
		// hex only 0-9a-f
		for _, ch := range tok {
			if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') {
				t.Fatalf("non-hex char in token: %q", tok)
			}
		}
	}
}
