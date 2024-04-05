package main

import (
	"github.com/Neeeooshka/alice-skill.git/internal/config"
	"github.com/thanhpk/randstr"
)

type shortener struct {
	options config.Options
	config  config.Config
}

func (s *shortener) GetBaseURL() string {
	baseURL := s.options.GetBaseURL()
	if s.config.BaseURL != "" {
		baseURL = s.config.BaseURL
	}

	return baseURL
}

func (s *shortener) GenerateShortLink() string {
	return s.GetBaseURL() + "/" + randstr.Base62(8)
}

func newShortener(opt config.Options, cfg config.Config) shortener {
	return shortener{options: opt, config: cfg}
}
