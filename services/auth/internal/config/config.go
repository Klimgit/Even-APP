package config

import (
	"fmt"
	"os"
	"time"

	libconfig "github.com/even-app/even-app/libs/config"
)

const DefaultHTTPPort = 8081

type Config struct {
	Base        libconfig.Base
	DatabaseURL string
	JWTSecret   string
	AccessTTL   time.Duration
	RefreshTTL  time.Duration
}

func Load() (Config, error) {
	base, err := libconfig.LoadBase(DefaultHTTPPort)
	if err != nil {
		return Config{}, err
	}
	dbURL, err := libconfig.MustGetenv("DATABASE_URL")
	if err != nil {
		return Config{}, err
	}
	jwt, err := libconfig.MustGetenv("JWT_SECRET")
	if err != nil {
		return Config{}, err
	}
	access, err := parseDuration("JWT_ACCESS_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, err
	}
	refresh, err := parseDuration("JWT_REFRESH_TTL", 7*24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	return Config{
		Base:        base,
		DatabaseURL: dbURL,
		JWTSecret:   jwt,
		AccessTTL:   access,
		RefreshTTL:  refresh,
	}, nil
}

func parseDuration(key string, def time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return d, nil
}

func (c Config) Validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	return nil
}
