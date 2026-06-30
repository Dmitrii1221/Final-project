package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"final-project/internal/config"
	budgetperiodrepo "final-project/internal/repository/budget_period"
	currencyrepo "final-project/internal/repository/currency"
	periodbalancerepo "final-project/internal/repository/period_balance"
	periodlimitrepo "final-project/internal/repository/period_limit"
	rolerepo "final-project/internal/repository/role"
	spendingrepo "final-project/internal/repository/spending"
	userbudgetrolerepo "final-project/internal/repository/user_budget_role"
	kafkatransport "final-project/internal/transport/kafka"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	slog.Info("starting budget-consumer")

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.PostgresDSN)
	if err != nil {
		slog.Error("cannot connect db", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	handler := kafkatransport.NewHandler(
		currencyrepo.NewPostgres(pool),
		budgetperiodrepo.NewPostgres(pool),
		periodlimitrepo.NewPostgres(pool),
		spendingrepo.NewPostgres(pool),
		periodbalancerepo.NewPostgres(pool),
		userbudgetrolerepo.NewPostgres(pool),
		rolerepo.NewPostgres(pool),
	)
	consumer := kafkatransport.New(
		cfg.KafkaBrokers,
		cfg.KafkaTopicSpendings,
		cfg.KafkaGroupID,
		cfg.KafkaTopicDLQ,
		handler,
	)

	if err := consumer.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
