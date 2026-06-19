package config

import (
	"os"
	"time"

	libconfig "github.com/even-app/even-app/libs/config"
)

const DefaultHTTPPort = 8083

type Config struct {
	Base                 libconfig.Base
	DatabaseURL          string
	LearningDatabaseURL  string
	AuthDatabaseURL      string
	LearningServiceURL   string
	AuthServiceURL       string
	InternalServiceToken string
	JWTSecret            string
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
		Base:                 base,
		DatabaseURL:          dbURL,
		LearningDatabaseURL:  os.Getenv("LEARNING_DATABASE_URL"),
		AuthDatabaseURL:      os.Getenv("AUTH_DATABASE_URL"),
		LearningServiceURL:   os.Getenv("LEARNING_SERVICE_URL"),
		AuthServiceURL:       os.Getenv("AUTH_SERVICE_URL"),
		InternalServiceToken: libconfig.InternalServiceToken(),
		JWTSecret:            jwt,
	}, nil
}

func (c Config) HasLearningHTTP() bool {
	return c.LearningServiceURL != ""
}

func (c Config) HasAuthHTTP() bool {
	return c.AuthServiceURL != ""
}

func (c Config) HasLearningDB() bool {
	return c.LearningDatabaseURL != ""
}

func (c Config) HasAuthDB() bool {
	return c.AuthDatabaseURL != ""
}

func (c Config) AccessTTL() time.Duration {
	return 15 * time.Minute
}
