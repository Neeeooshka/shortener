package main

import (
	"flag"
	"github.com/Neeeooshka/alice-skill.git/internal/config"
	"github.com/Neeeooshka/alice-skill.git/internal/gzip"
	"github.com/Neeeooshka/alice-skill.git/internal/handlers"
	"github.com/Neeeooshka/alice-skill.git/internal/zap"
	"github.com/Neeeooshka/alice-skill.git/pkg/compressor"
	logger2 "github.com/Neeeooshka/alice-skill.git/pkg/logger"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func main() {
	opt := getOptions()

	zapLoger, err := zap.NewZapLogger("info")

	if err != nil {
		panic(err)
	}

	sh := newShortener(opt)

	// create router
	router := chi.NewRouter()
	router.Post("/", logger2.IncludeLogger(compressor.IncludeCompressor(handlers.GetShortenerHandler(&sh), gzip.NewGzipCompressor()), zapLoger))
	router.Post("/api/shorten", logger2.IncludeLogger(compressor.IncludeCompressor(handlers.GetAPIShortenHandler(&sh), gzip.NewGzipCompressor()), zapLoger))
	router.Get("/{id}", logger2.IncludeLogger(handlers.GetExpanderHandler(&sh), zapLoger))

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
