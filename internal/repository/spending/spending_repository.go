package spendingrepo

import (
	"context"
	"final-project/internal/domain"
	"fmt"
	"time"

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
func (r *PostgresRepo) ListByBudgetID(ctx context.Context, budgetID int64, currencyID *int64, periodID *int64, from, to *time.Time) ([]domain.Spending, error) {
	builder := psql.Select("id", "period_id", "budget_id", "currency_id", "amount", "idempotency_key", "spent_at", "created_at").
		From("spendings").
		Where("budget_id = ?", budgetID)
	if currencyID != nil {
		builder = builder.Where("currency_id = ?", *currencyID)
	}
	if periodID != nil {
		builder = builder.Where("period_id = ?", *periodID)
	}
	if from != nil {
		builder = builder.Where("spent_at >= ?", *from)
	}
	if to != nil {
		builder = builder.Where("spent_at <= ?", *to)
	}
	builder = builder.OrderBy("spent_at ASC")

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select: %w", err)
	}
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("exec select: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Spending, 0)
	for rows.Next() {
		var s domain.Spending
		if err := rows.Scan(
			&s.ID, &s.PeriodID, &s.BudgetID, &s.CurrencyID,
			&s.Amount, &s.IdempotencyKey, &s.SpentAt, &s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("Scan: %w", err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}
