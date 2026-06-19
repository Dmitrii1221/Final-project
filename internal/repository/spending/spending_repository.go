package spendingrepo

import (
	"context"
	"final-project/internal/domain"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type PostgresRepo struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{pool: pool}
}

func (r *PostgresRepo) Insert(ctx context.Context, s domain.Spending) (domain.Spending, error) {
	query, args, err := psql.
		Insert("spendings").
		Columns("period_id", "budget_id", "currency_id", "amount", "idempotency_key", "spent_at").
		Values(s.PeriodID, s.BudgetID, s.CurrencyID, s.Amount, s.IdempotencyKey, s.SpentAt).
		Suffix("RETURNING id, budget_id, period_id, currency_id, amount, idempotency_key, spent_at, created_at").
		ToSql()
	if err != nil {
		return domain.Spending{}, fmt.Errorf("Build insert: %w", err)
	}

	var spend domain.Spending
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&spend.ID,
		&spend.PeriodID,
		&spend.BudgetID,
		&spend.CurrencyID,
		&spend.Amount,
		&spend.IdempotencyKey,
		&spend.SpentAt,
		&spend.CreatedAt,
	)
	if err != nil {
		return domain.Spending{}, fmt.Errorf("exec insert: %w", err)
	}
	return spend, nil
}

func (r *PostgresRepo) GetByID(ctx context.Context, id int64) (domain.Spending, error) {
	query, args, err := psql.
		Select("id", "period_id", "budget_id", "currency_id", "amount", "idempotency_key", "spent_at", "created_at").
		From("spendings").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return domain.Spending{}, fmt.Errorf("build query: %w", err)
	}

	var spend domain.Spending
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&spend.ID,
		&spend.PeriodID,
		&spend.BudgetID,
		&spend.CurrencyID,
		&spend.Amount,
		&spend.IdempotencyKey,
		&spend.SpentAt,
		&spend.CreatedAt,
	)
	if err != nil {
		return domain.Spending{}, fmt.Errorf("spend %d: %w", id, err)
	}
	return spend, nil
}

func (r *PostgresRepo) GetByIdempotencyKey(ctx context.Context, budgetID int64, key string) (domain.Spending, error) {
	query, args, err := psql.
		Select("id", "period_id", "budget_id", "currency_id", "amount", "idempotency_key", "spent_at", "created_at").
		From("spendings").
		Where("idempotency_key = ? AND budget_id = ?", key, budgetID).
		ToSql()
	if err != nil {
		return domain.Spending{}, fmt.Errorf("build select: %w", err)
	}

	var spend domain.Spending
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&spend.ID,
		&spend.PeriodID,
		&spend.BudgetID,
		&spend.CurrencyID,
		&spend.Amount,
		&spend.IdempotencyKey,
		&spend.SpentAt,
		&spend.CreatedAt,
	)
	if err != nil {
		return domain.Spending{}, fmt.Errorf("idempotency key %s: %w", key, err)
	}
	return spend, nil
}
