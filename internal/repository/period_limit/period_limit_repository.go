package periodlimitrepo

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

func (r *PostgresRepo) Create(ctx context.Context, p domain.PeriodLimit) (domain.PeriodLimit, error) {
	query, args, err := psql.
		Insert("period_limits").
		Columns("period_id", "currency_id", "limit_amount").
		Values(p.PeriodID, p.CurrencyID, p.LimitAmount).
		Suffix("RETURNING id, period_id, currency_id, limit_amount").
		ToSql()
	if err != nil {
		return domain.PeriodLimit{}, fmt.Errorf("Build insert: %w", err)
	}

	var period domain.PeriodLimit
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&period.ID,
		&period.PeriodID,
		&period.CurrencyID,
		&period.LimitAmount,
	)
	if err != nil {
		return domain.PeriodLimit{}, fmt.Errorf("exec insert: %w", err)
	}
	return period, nil
}

func (r *PostgresRepo) GetByPeriodAndCurrency(ctx context.Context, periodID int64, currencyID int64) (domain.PeriodLimit, error) {
	query, args, err := psql.
		Select("id", "period_id", "currency_id", "limit_amount").
		From("period_limits").
		Where("period_id = ? AND currency_id = ?", periodID, currencyID).
		ToSql()
	if err != nil {
		return domain.PeriodLimit{}, fmt.Errorf("build query: %w", err)
	}

	var period domain.PeriodLimit
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&period.ID,
		&period.PeriodID,
		&period.CurrencyID,
		&period.LimitAmount,
	)
	if err != nil {
		return domain.PeriodLimit{}, fmt.Errorf("period %d: %w", periodID, err)
	}
	return period, nil
}

func (r *PostgresRepo) ListByPeriodID(ctx context.Context, periodID int64) ([]domain.PeriodLimit, error) {
	query, args, err := psql.
		Select("id", "period_id", "currency_id", "limit_amount").
		From("period_limits").
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

	out := make([]domain.PeriodLimit, 0)
	for rows.Next() {
		var period domain.PeriodLimit
		if err := rows.Scan(
			&period.ID,
			&period.PeriodID,
			&period.CurrencyID,
			&period.LimitAmount,
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
