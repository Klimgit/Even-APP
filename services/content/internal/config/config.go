package config

import (
	libconfig "github.com/even-app/even-app/libs/config"
)

const DefaultHTTPPort = 8083

type Config struct {
	Base        libconfig.Base
	DatabaseURL string
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
	return Config{Base: base, DatabaseURL: dbURL}, nil
}
