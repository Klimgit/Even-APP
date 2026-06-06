package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Base holds common env-backed settings for all services.
type Base struct {
	HTTPPort        int
	LogLevel        string
	ShutdownTimeout time.Duration
}

// LoadBase reads HTTP_PORT, LOG_LEVEL, SHUTDOWN_TIMEOUT.
func LoadBase(defaultPort int) (Base, error) {
	port, err := envInt("HTTP_PORT", defaultPort)
	if err != nil {
		return Base{}, err
	}
	shutdown, err := envDuration("SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Base{}, err
	}
	return Base{
		HTTPPort:        port,
		LogLevel:        envString("LOG_LEVEL", "info"),
		ShutdownTimeout: shutdown,
	}, nil
}

// MustGetenv returns an error if the variable is unset or empty.
func MustGetenv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("required env %s is not set", key)
	}
	return v, nil
}

func envString(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return n, nil
}

func envDuration(key string, def time.Duration) (time.Duration, error) {
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
