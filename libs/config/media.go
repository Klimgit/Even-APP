package config

import (
	"fmt"
	"os"
	"strconv"
)

// Media holds storage quota settings for per-user media libraries.
type Media struct {
	// UserQuotaBytes is the max total size of active media per user (0 = unlimited).
	UserQuotaBytes int64
}

// LoadMedia reads MEDIA_USER_QUOTA_BYTES (bytes). Default 524288000 (500 MiB).
func LoadMedia() Media {
	raw := os.Getenv("MEDIA_USER_QUOTA_BYTES")
	if raw == "" {
		return Media{UserQuotaBytes: 524288000}
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n < 0 {
		return Media{UserQuotaBytes: 524288000}
	}
	return Media{UserQuotaBytes: n}
}

func (m Media) Validate() error {
	if m.UserQuotaBytes < 0 {
		return fmt.Errorf("MEDIA_USER_QUOTA_BYTES must be >= 0")
	}
	return nil
}
