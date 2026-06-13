package main

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"final-project/internal/config"
	"final-project/internal/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/pressly/goose/v3"
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

	db, err := sql.Open("pgx", cfg.PostgresDSN)
	if err != nil {
		slog.Error("cannot open db", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	goose.SetDialect("postgres")
	if err := goose.Up(db, "migrations"); err != nil {
		slog.Error("migrations failed", "err", err)
		os.Exit(1)
	}

	slog.Info("migrations applied successfully")

	if err := e.Start(cfg.HTTPAddr); err != nil {
		slog.Error("Ошибка запуска сервера:", "err", err)
		os.Exit(1)
	}

}
