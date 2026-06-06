package media

import (
	"strings"
	"testing"
	"time"
)

func TestKindFromMIME(t *testing.T) {
	tests := []struct {
		mime    string
		want    string
		wantErr bool
	}{
		{"image/png", "image", false},
		{"IMAGE/JPEG", "image", false},
		{"audio/mpeg", "audio", false},
		{"video/mp4", "video", false},
		{"application/pdf", "", true},
	}
	for _, tt := range tests {
		got, err := KindFromMIME(tt.mime)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("KindFromMIME(%q) expected error", tt.mime)
			}
			continue
		}
		if err != nil || got != tt.want {
			t.Fatalf("KindFromMIME(%q) = %q, %v; want %q", tt.mime, got, err, tt.want)
		}
	}
}

func TestValidateDisplayName(t *testing.T) {
	if err := ValidateDisplayName("  hello  "); err != nil {
		t.Fatalf("valid name: %v", err)
	}
	if err := ValidateDisplayName(""); err == nil {
		t.Fatal("empty name should fail")
	}
	if err := ValidateDisplayName(strings.Repeat("a", 121)); err == nil {
		t.Fatal("long name should fail")
	}
}

func TestResolveExpires(t *testing.T) {
	ttl := int64(3600)
	exp, err := ResolveExpires(&ttl, nil)
	if err != nil || exp == nil || !exp.After(time.Now().UTC()) {
		t.Fatalf("ttl: exp=%v err=%v", exp, err)
	}

	future := time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)
	exp, err = ResolveExpires(nil, &future)
	if err != nil || exp == nil {
		t.Fatalf("expires_at: %v", err)
	}

	exp, err = ResolveExpires(nil, nil)
	if err != nil || exp != nil {
		t.Fatalf("permanent: exp=%v err=%v", exp, err)
	}

	past := time.Now().UTC().Add(-time.Hour).Format(time.RFC3339)
	if _, err := ResolveExpires(nil, &past); err == nil {
		t.Fatal("past expires_at should fail")
	}
}
