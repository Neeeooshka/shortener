package main

import (
	"flag"
	"github.com/Neeeooshka/alice-skill.git/internal/config"
	"github.com/Neeeooshka/alice-skill.git/internal/storage"
	file "github.com/Neeeooshka/alice-skill.git/internal/storage/file"
	postgres "github.com/Neeeooshka/alice-skill.git/internal/storage/postgres"
	"github.com/Neeeooshka/alice-skill.git/pkg/compressor"
	"github.com/Neeeooshka/alice-skill.git/pkg/compressor/gzip"
	"github.com/Neeeooshka/alice-skill.git/pkg/logger"
	"github.com/Neeeooshka/alice-skill.git/pkg/logger/zap"
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

	var store storage.LinkStorage
	if opt.DB.String() != "" {
		store, err = postgres.NewPostgresLinksStorage(opt.DB.String())
		if err != nil {
			panic(err)
		}
	} else {
		store, _ = file.NewFileLinksStorage(opt.FileStorage.String())
	}
	defer store.Close()

	appInstance := newAppInstance(opt, store)

	// create router
	router := chi.NewRouter()
	router.Post("/", logger.IncludeLogger(compressor.IncludeCompressor(appInstance.ShortenerHandler, gzip.NewGzipCompressor()), zapLoger))
	router.Post("/api/shorten", logger.IncludeLogger(compressor.IncludeCompressor(appInstance.APIShortenerHandler, gzip.NewGzipCompressor()), zapLoger))
	router.Get("/{id}", logger.IncludeLogger(appInstance.ExpanderHandler, zapLoger))
	router.Get("/ping", logger.IncludeLogger(store.PingHandler, zapLoger))

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
	flag.Var(&opt.DB, "d", "postrgres connection string")

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

	if cfg.DB != "" {
		opt.DB.Set(cfg.DB)
	}

	return opt
}
