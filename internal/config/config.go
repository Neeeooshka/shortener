package config

import (
	"github.com/caarlos0/env/v6"
)

var cfg config

type config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

func init() {
	env.Parse(&cfg)
}

func GetConfig() *config {
	return &cfg
}
