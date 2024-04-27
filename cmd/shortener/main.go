package main

import (
	"flag"
	"github.com/Neeeooshka/alice-skill.git/internal/compress"
	"github.com/Neeeooshka/alice-skill.git/internal/config"
	"github.com/Neeeooshka/alice-skill.git/internal/handlers"
	"github.com/Neeeooshka/alice-skill.git/internal/logger"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	opt := getOptions()

	logrusLogger := newLogrusLogger(logrus.InfoLevel)

	sh := newShortener(opt)

	// create router
	router := chi.NewRouter()
	router.Post("/", logger.IncludeLogger(compress.IncludeCompressor(handlers.GetShortenerHandler(&sh), newGzipCompressor()), logrusLogger))
	router.Post("/api/shorten", logger.IncludeLogger(compress.IncludeCompressor(handlers.GetAPIShortenHandler(&sh), newGzipCompressor()), logrusLogger))
	router.Get("/{id}", logger.IncludeLogger(handlers.GetExpanderHandler(&sh), logrusLogger))

	// create HTTP Server
	http.ListenAndServe(opt.GetServer(), router)
}

// init options
func getOptions() config.Options {
	opt := config.NewOptions()
	cfg := config.NewConfig()

	flag.Var(&opt.ServerAddress, "a", "Server address - host:port")
	flag.Var(&opt.BaseURL, "b", "Server ShortLink Base address - protocol://host:port")
	flag.Var(&opt.FileStorage, "f", "File storage path for shortlinks")

	flag.Parse()
	env.Parse(&cfg)

	if cfg.ServerAddress != "" {
		opt.ServerAddress.Set(cfg.ServerAddress)
	}

	if cfg.BaseURL != "" {
		opt.BaseURL.Set(cfg.BaseURL)
	}

	if cfg.FileStorage != "" {
		opt.FileStorage.Set(cfg.FileStorage)
	}

	return opt
}
