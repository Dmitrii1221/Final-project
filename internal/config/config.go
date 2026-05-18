package config

import (
	"os"
	"time"
)

type ConfigServer struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func LoadFromEnv() *ConfigServer {
	return &ConfigServer{
		Port:         getEnv("HTTP_PORT", "1323"),
		ReadTimeout:  getDurationEnv("READ_TIMEOUT", 30*time.Second),
		WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 30*time.Second),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if durationVal, err := time.ParseDuration(value); err == nil {
			return durationVal
		}
	}
	return defaultValue
}
