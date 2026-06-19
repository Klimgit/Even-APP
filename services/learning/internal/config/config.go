package config

import (
	"os"
	"time"

	libconfig "github.com/even-app/even-app/libs/config"
)

const DefaultHTTPPort = 8084

type Config struct {
	Base                 libconfig.Base
	DatabaseURL          string
	ContentDatabaseURL   string
	LexiconDatabaseURL   string
	MediaDatabaseURL     string
	ContentServiceURL    string
	LexiconServiceURL    string
	MediaServiceURL      string
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
	contentURL := os.Getenv("CONTENT_DATABASE_URL")
	lexiconURL := os.Getenv("LEXICON_DATABASE_URL")
	mediaURL := os.Getenv("MEDIA_DATABASE_URL")
	return Config{
		Base:                 base,
		DatabaseURL:          dbURL,
		ContentDatabaseURL:   contentURL,
		LexiconDatabaseURL:   lexiconURL,
		MediaDatabaseURL:     mediaURL,
		ContentServiceURL:    os.Getenv("CONTENT_SERVICE_URL"),
		LexiconServiceURL:    os.Getenv("LEXICON_SERVICE_URL"),
		MediaServiceURL:      os.Getenv("MEDIA_SERVICE_URL"),
		InternalServiceToken: libconfig.InternalServiceToken(),
		JWTSecret:            jwt,
	}, nil
}

func (c Config) HasContentHTTP() bool {
	return c.ContentServiceURL != ""
}

func (c Config) HasLexiconHTTP() bool {
	return c.LexiconServiceURL != ""
}

func (c Config) HasMediaHTTP() bool {
	return c.MediaServiceURL != ""
}

func (c Config) HasMediaDB() bool {
	return c.MediaDatabaseURL != ""
}

func (c Config) HasLexiconDB() bool {
	return c.LexiconDatabaseURL != ""
}

func (c Config) AccessTTL() time.Duration {
	return 15 * time.Minute
}

func (c Config) HasContentDB() bool {
	return c.ContentDatabaseURL != ""
}
