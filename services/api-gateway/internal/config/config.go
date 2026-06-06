package config

import (
	libconfig "github.com/even-app/even-app/libs/config"
)

const DefaultHTTPPort = 8080

type Config struct {
	Base        libconfig.Base
	JWTSecret   string
	AuthURL     string
	LexiconURL  string
	ContentURL  string
	LearningURL string
}

func Load() (Config, error) {
	base, err := libconfig.LoadBase(DefaultHTTPPort)
	if err != nil {
		return Config{}, err
	}
	jwt, err := libconfig.MustGetenv("JWT_SECRET")
	if err != nil {
		return Config{}, err
	}
	auth, err := libconfig.MustGetenv("AUTH_URL")
	if err != nil {
		return Config{}, err
	}
	lexicon, err := libconfig.MustGetenv("LEXICON_URL")
	if err != nil {
		return Config{}, err
	}
	content, err := libconfig.MustGetenv("CONTENT_URL")
	if err != nil {
		return Config{}, err
	}
	learning, err := libconfig.MustGetenv("LEARNING_URL")
	if err != nil {
		return Config{}, err
	}
	return Config{
		Base:        base,
		JWTSecret:   jwt,
		AuthURL:     auth,
		LexiconURL:  lexicon,
		ContentURL:  content,
		LearningURL: learning,
	}, nil
}
