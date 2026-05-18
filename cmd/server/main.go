package main

import (
	"log"
	"net/http"

	"final-project/internal/config"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("Нет .env файла")
	}

	cfg := config.LoadFromEnv()

	e := echo.New()

	// Middleware для Echo v4
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Health check
	e.GET("/healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "Working")
	})

	addr := ":" + cfg.Port
	log.Printf("Сервер запустился на порту: %s", addr)

	if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
