package config

import (
	"fmt"

	libconfig "github.com/even-app/even-app/libs/config"
)

const DefaultHTTPPort = 8081

type Config struct {
	Base        libconfig.Base
	DatabaseURL string
	JWTSecret   string
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
	return Config{
		Base:        base,
		DatabaseURL: dbURL,
		JWTSecret:   jwt,
	}, nil
}

func (c Config) Validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	return nil
}
