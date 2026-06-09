package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	HTTPAddr    string `env:"HTTP_ADDR" envDefault:":8080"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
	Env         string `env:"APP_ENV"   envDefault:"local"` // local | dev | prod
	PostgresDSN string `env:"POSTGRES_DSN,required"`
}

func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}
