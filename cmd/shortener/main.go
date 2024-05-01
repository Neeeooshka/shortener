package main

import (
	"flag"
	"github.com/Neeeooshka/alice-skill.git/internal/config"
	"github.com/Neeeooshka/alice-skill.git/internal/gzip"
	"github.com/Neeeooshka/alice-skill.git/internal/handlers"
	"github.com/Neeeooshka/alice-skill.git/internal/logger"
	"github.com/Neeeooshka/alice-skill.git/pkg/compressor"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
)

type zapLogger struct {
	logger *zap.Logger
}

func (l *zapLogger) Log(rq logger.RequestData, rs logger.ResponseData) {
	l.logger.Debug("receive new request",
		zap.String("URI", rq.URI),
		zap.String("method", rq.Method),
		zap.Duration("duration", rq.Duration),
		zap.Int("status", rs.Status),
		zap.Int("size", rs.Size),
	)
}

func main() {
	opt := getOptions()

	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}

	zapLoger := &zapLogger{logger: logger.Log}

	sh := newShortener(opt)

	// create router
	router := chi.NewRouter()
	router.Post("/", logger.IncludeLogger(compressor.IncludeCompressor(handlers.GetShortenerHandler(&sh), gzip.NewGzipCompressor()), zapLoger))
	router.Post("/api/shorten", logger.IncludeLogger(compressor.IncludeCompressor(handlers.GetAPIShortenHandler(&sh), gzip.NewGzipCompressor()), zapLoger))
	router.Get("/{id}", logger.IncludeLogger(handlers.GetExpanderHandler(&sh), zapLoger))

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
