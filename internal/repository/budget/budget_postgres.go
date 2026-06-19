package budgetrepo

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

func (r *PostgresRepo) Insert(ctx context.Context, budget domain.Budget) (domain.Budget, error) {
	query, args, err := psql.
		Insert("budgets").
		Columns("name", "owner_user_id").
		Values(budget.Name, budget.OwnerUserID).
		Suffix("RETURNING id, name,created_at, updated_at, owner_user_id").
		ToSql()
	if err != nil {
		return domain.Budget{}, fmt.Errorf("Build insert: %w", err)
	}

	var b domain.Budget
	err = r.pool.QueryRow(ctx, query, args...).Scan(&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt, &b.OwnerUserID)
	if err != nil {
		return domain.Budget{}, fmt.Errorf("exec insert: %w", err)
	}
	return b, nil
}

func (r *PostgresRepo) List(ctx context.Context) ([]domain.Budget, error) {
	query, args, err := psql.
		Select("id", "name", "created_at", "updated_at", "owner_user_id").
		From("budgets").
		OrderBy("id ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("exec select: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Budget, 0)
	for rows.Next() {
		var b domain.Budget
		if err := rows.Scan(&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt, &b.OwnerUserID); err != nil {
			return nil, fmt.Errorf("Scan: %w", err)
		}
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

func (r *PostgresRepo) GetByID(ctx context.Context, id int64) (domain.Budget, error) {
	query, args, err := psql.
		Select("id", "name", "created_at", "updated_at", "owner_user_id").
		From("budgets").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return domain.Budget{}, fmt.Errorf("build select: %w", err)
	}

	var b domain.Budget
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt, &b.OwnerUserID,
	)
	if err != nil {
		return domain.Budget{}, fmt.Errorf("budget %d: %w", id, err)
	}
	return b, nil
}
