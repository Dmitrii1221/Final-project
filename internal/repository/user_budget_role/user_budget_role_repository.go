package userbudgetrolerepo

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

func (r *PostgresRepo) Grant(ctx context.Context, ubr domain.UserBudgetRole) error {
	query, args, err := psql.
		Insert("user_budget_roles").
		Columns("user_id", "budget_id", "role_id").
		Values(ubr.UserID, ubr.BudgetID, ubr.RoleID).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec insert: %w", err)
	}
	return nil
}

func (r *PostgresRepo) Revoke(ctx context.Context, userID, budgetID, roleID int64) error {
	query, args, err := psql.
		Delete("user_budget_roles").
		Where("user_id = ? AND budget_id = ? AND role_id = ?", userID, budgetID, roleID).
		ToSql()
	if err != nil {
		return fmt.Errorf("build delete: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec delete: %w", err)
	}
	return nil
}

func (r *PostgresRepo) GetByUserAndBudget(ctx context.Context, userID, budgetID int64) ([]domain.UserBudgetRole, error) {
	query, args, err := psql.
		Select("user_id", "budget_id", "role_id").
		From("user_budget_roles").
		Where("user_id = ? AND budget_id = ?", userID, budgetID).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("exec select: %w", err)
	}
	defer rows.Close()

	out := make([]domain.UserBudgetRole, 0)
	for rows.Next() {
		var ubr domain.UserBudgetRole
		if err := rows.Scan(&ubr.UserID, &ubr.BudgetID, &ubr.RoleID); err != nil {
			return nil, fmt.Errorf("Scan: %w", err)
		}
		out = append(out, ubr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

func (r *PostgresRepo) GetByUserID(ctx context.Context, userID int64) ([]domain.UserBudgetRole, error) {
	query, args, err := psql.
		Select("user_id", "budget_id", "role_id").
		From("user_budget_roles").
		Where("user_id = ?", userID).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("exec select: %w", err)
	}
	defer rows.Close()

	out := make([]domain.UserBudgetRole, 0)
	for rows.Next() {
		var ubr domain.UserBudgetRole
		if err := rows.Scan(&ubr.UserID, &ubr.BudgetID, &ubr.RoleID); err != nil {
			return nil, fmt.Errorf("Scan: %w", err)
		}
		out = append(out, ubr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}
