package config

import (
	"time"

	libconfig "github.com/even-app/even-app/libs/config"
)

const DefaultHTTPPort = 8085

type Config struct {
	Base        libconfig.Base
	DatabaseURL string
	S3          libconfig.S3
	Media       libconfig.Media
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
	s3, err := libconfig.LoadS3()
	if err != nil {
		return Config{}, err
	}
	jwt, err := libconfig.MustGetenv("JWT_SECRET")
	if err != nil {
		return Config{}, err
	}
	return Config{Base: base, DatabaseURL: dbURL, S3: s3, Media: libconfig.LoadMedia(), JWTSecret: jwt}, nil
}

func (c Config) AccessTTL() time.Duration {
	return 15 * time.Minute
}
