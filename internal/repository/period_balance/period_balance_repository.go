package periodbalancerepo

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

func (r *PostgresRepo) Upsert(ctx context.Context, p domain.PeriodBalance) (domain.PeriodBalance, error) {
	query, args, err := psql.
		Insert("period_balances").
		Columns("period_id", "currency_id", "remaining").
		Values(p.PeriodID, p.CurrencyID, p.Remaining).
		Suffix(`ON CONFLICT (period_id, currency_id) 
                DO UPDATE SET remaining = EXCLUDED.remaining, updated_at = now()
                RETURNING period_id, currency_id, remaining, updated_at`).
		ToSql()
	if err != nil {
		return domain.PeriodBalance{}, fmt.Errorf("Build insert: %w", err)
	}

	var period domain.PeriodBalance
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&period.PeriodID,
		&period.CurrencyID,
		&period.Remaining,
		&period.UpdatedAt,
	)
	if err != nil {
		return domain.PeriodBalance{}, fmt.Errorf("exec insert: %w", err)
	}
	return period, nil
}

func (r *PostgresRepo) GetByPeriodAndCurrency(ctx context.Context, periodID int64, currencyID int64) (domain.PeriodBalance, error) {
	query, args, err := psql.
		Select("period_id", "currency_id", "remaining", "updated_at").
		From("period_balances").
		Where("period_id = ? AND currency_id = ?", periodID, currencyID).
		ToSql()
	if err != nil {
		return domain.PeriodBalance{}, fmt.Errorf("build query: %w", err)
	}

	var period domain.PeriodBalance
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&period.PeriodID,
		&period.CurrencyID,
		&period.Remaining,
		&period.UpdatedAt,
	)
	if err != nil {
		return domain.PeriodBalance{}, fmt.Errorf("period %d: %w", periodID, err)
	}
	return period, nil
}

func (r *PostgresRepo) ListByPeriodID(ctx context.Context, periodID int64) ([]domain.PeriodBalance, error) {
	query, args, err := psql.
		Select("period_id", "currency_id", "remaining", "updated_at").
		From("period_balances").
		Where("period_id = ?", periodID).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("exec select: %w", err)
	}
	defer rows.Close()

	out := make([]domain.PeriodBalance, 0)
	for rows.Next() {
		var period domain.PeriodBalance
		if err := rows.Scan(
			&period.PeriodID,
			&period.CurrencyID,
			&period.Remaining,
			&period.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("Scan: %w", err)
		}
		out = append(out, period)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}
