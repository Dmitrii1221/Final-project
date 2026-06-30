package config

import (
	"fmt"
	"log"
	"time"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr     string        `env:"HTTP_ADDR" envDefault:":8080"`
	LogLevel     string        `env:"LOG_LEVEL" envDefault:"info"`
	Env          string        `env:"APP_ENV"   envDefault:"local"`
	PostgresDSN  string        `env:"POSTGRES_DSN,required"`
	JWTSecret    string        `env:"JWT_SECRET,required"`
	JWTAccessTTL time.Duration `env:"JWT_ACCESS_TTL" envDefault:"15m"`
	GRPCAddr            string `env:"GRPC_ADDR" envDefault:":9090"`
	KafkaBrokers        string `env:"KAFKA_BROKERS"        envDefault:"localhost:9092"`
	KafkaTopicSpendings string `env:"KAFKA_TOPIC_SPENDINGS" envDefault:"spendings"`
	KafkaTopicDLQ       string `env:"KAFKA_TOPIC_DLQ"       envDefault:"spendings.dlq"`
	KafkaGroupID        string `env:"KAFKA_GROUP_ID"        envDefault:"budget-consumer"`
}

func Load() (Config, error) {
	// Загружаем .env файл в системные переменные
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}
