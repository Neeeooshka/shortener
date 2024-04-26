package main

import (
	"github.com/Neeeooshka/alice-skill.git/internal/config"
	"github.com/Neeeooshka/alice-skill.git/internal/storage"
	"github.com/thanhpk/randstr"
)

type shortener struct {
	options config.Options
	storage storage.Links
}

func (s *shortener) GetBaseURL() string {
	return s.options.GetBaseURL()
}

func (s *shortener) GenerateShortLink() string {
	return s.GetBaseURL() + "/" + randstr.Base62(8)
}

func (s *shortener) Add(sl, fl string) { s.storage.Add(sl, fl) }

func (s *shortener) Get(shortLink string) (string, bool) { return s.storage.Get(shortLink) }

func newShortener(opt config.Options) shortener {
	return shortener{options: opt, storage: *storage.NewLinksStorage(opt.GetFileStorage())}
}
