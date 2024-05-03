package config

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
	FileStorage   string `env:"FILE_STORAGE_PATH"`
	DB            string `env:"DATABASE_DSN"`
}

func NewConfig() Config {
	return Config{}
}
