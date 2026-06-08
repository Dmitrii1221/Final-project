package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"final-project/internal/config"
	"final-project/internal/logger"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("Нет .env файла")
	}

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	e := echo.New()
	e.HideBanner = true

	log := logger.New(cfg.LogLevel, cfg.Env)
	slog.SetDefault(log)

	log.Info("starting http server", "addr", cfg.HTTPAddr, "env", cfg.Env)

	// Health check
	e.GET("/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			log.Info("http request",
				"method", c.Request().Method,
				"path", c.Path(),
				"status", c.Response().Status,
				"dur_ms", time.Since(start).Microseconds(),
			)
			return err
		}
	})

	if err := e.Start(cfg.HTTPAddr); err != nil {
		log.Error("Ошибка запуска сервера:", "err", err)
		os.Exit(1)
	}

}
