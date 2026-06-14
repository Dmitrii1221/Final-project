package main

import (
	"context"
	"final-project/internal/config"
	"final-project/internal/domain"
	"final-project/internal/logger"
	budgetrepo "final-project/internal/repository/budget"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log := logger.New(cfg.LogLevel, cfg.Env)
	slog.SetDefault(log)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Error("connect postgres", "err: ", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Error("ping postgres", "err: ", err)
		os.Exit(1)
	}

	repo := budgetrepo.NewPostgres(pool)

	for _, name := range []string{"alpha", "beta", "gamma"} {
		b, err := repo.Insert(ctx, domain.Budget{Name: name})
		if err != nil {
			log.Error("insert", "name", name, "err", err)
			os.Exit(1)
		}
		log.Info("inserted", "id", b.ID, "name", b.Name)
	}

	budgets, err := repo.List(ctx)
	if err != nil {
		log.Error("list", "err", err)
		os.Exit(1)
	}
	for _, b := range budgets {
		log.Info("budget", "id", b.ID, "name", b.Name, "created_at", b.CreatedAt)
	}
}
