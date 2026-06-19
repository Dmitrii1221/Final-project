package budgetperiodrepo

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

func (r *PostgresRepo) Create(ctx context.Context, p domain.BudgetPeriod) (domain.BudgetPeriod, error) {
	query, args, err := psql.
		Insert("budget_periods").
		Columns("budget_id", "period_start", "period_end").
		Values(p.BudgetID, p.PeriodStart, p.PeriodEnd).
		Suffix("RETURNING id, budget_id, period_start, period_end, created_at, updated_at").
		ToSql()
	if err != nil {
		return domain.BudgetPeriod{}, fmt.Errorf("Build insert: %w", err)
	}

	var period domain.BudgetPeriod
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&period.ID,
		&period.BudgetID,
		&period.PeriodStart,
		&period.PeriodEnd,
		&period.CreatedAt,
		&period.UpdatedAt,
	)
	if err != nil {
		return domain.BudgetPeriod{}, fmt.Errorf("exec insert: %w", err)
	}
	return period, nil
}

func (r *PostgresRepo) GetByID(ctx context.Context, id int64) (domain.BudgetPeriod, error) {
	query, args, err := psql.
		Select("id", "budget_id", "period_start", "period_end", "created_at", "updated_at").
		From("budget_periods").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return domain.BudgetPeriod{}, fmt.Errorf("build query: %w", err)
	}

	var period domain.BudgetPeriod
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&period.ID,
		&period.BudgetID,
		&period.PeriodStart,
		&period.PeriodEnd,
		&period.CreatedAt,
		&period.UpdatedAt,
	)
	if err != nil {
		return domain.BudgetPeriod{}, fmt.Errorf("budget period %d: %w", id, err)
	}
	return period, nil
}

func (r *PostgresRepo) ListByBudgetID(ctx context.Context, budgetID int64) ([]domain.BudgetPeriod, error) {
	query, args, err := psql.
		Select("id", "budget_id", "period_start", "period_end", "created_at", "updated_at").
		From("budget_periods").
		Where("budget_id = ?", budgetID).
		OrderBy("period_start ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("exec select: %w", err)
	}
	defer rows.Close()

	out := make([]domain.BudgetPeriod, 0)
	for rows.Next() {
		var period domain.BudgetPeriod
		if err := rows.Scan(
			&period.ID,
			&period.BudgetID,
			&period.PeriodStart,
			&period.PeriodEnd,
			&period.CreatedAt,
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
