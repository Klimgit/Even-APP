package config

import (
	"os"
	"time"

	libconfig "github.com/even-app/even-app/libs/config"
)

const DefaultHTTPPort = 8082

type Config struct {
	Base               libconfig.Base
	DatabaseURL        string
	ContentDatabaseURL string
	JWTSecret          string
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
		Base:               base,
		DatabaseURL:        dbURL,
		ContentDatabaseURL: os.Getenv("CONTENT_DATABASE_URL"),
		JWTSecret:          jwt,
	}, nil
}

func (c Config) HasContentDB() bool {
	return c.ContentDatabaseURL != ""
}

func (c Config) AccessTTL() time.Duration {
	return 15 * time.Minute
}
