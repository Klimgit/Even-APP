package media

import (
	"fmt"
	"strings"
	"time"
)

func KindFromMIME(mime string) (string, error) {
	m := strings.ToLower(mime)
	switch {
	case strings.HasPrefix(m, "image/"):
		return "image", nil
	case strings.HasPrefix(m, "audio/"):
		return "audio", nil
	case strings.HasPrefix(m, "video/"):
		return "video", nil
	default:
		return "", fmt.Errorf("unsupported mime type: %s", mime)
	}
}

func ValidateDisplayName(name string) error {
	n := strings.TrimSpace(name)
	if n == "" {
		return fmt.Errorf("display_name is required")
	}
	if len(n) > 120 {
		return fmt.Errorf("display_name must be at most 120 characters")
	}
	return nil
}

// ResolveExpires returns nil for permanent storage. Accepts ttl_seconds or expires_at (RFC3339).
func ResolveExpires(ttlSeconds *int64, expiresAt *string) (*time.Time, error) {
	if ttlSeconds != nil && *ttlSeconds > 0 {
		t := time.Now().UTC().Add(time.Duration(*ttlSeconds) * time.Second)
		return &t, nil
	}
	if expiresAt != nil && strings.TrimSpace(*expiresAt) != "" {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(*expiresAt))
		if err != nil {
			return nil, fmt.Errorf("invalid expires_at")
		}
		if !t.After(time.Now().UTC()) {
			return nil, fmt.Errorf("expires_at must be in the future")
		}
		return &t, nil
	}
	return nil, nil
}
