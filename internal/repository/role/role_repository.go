package rolerepo

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

func (r *PostgresRepo) Insert(ctx context.Context, c domain.Role) (domain.Role, error) {
	query, args, err := psql.
		Insert("roles").
		Columns("code", "description").
		Values(c.Code, c.Description).
		Suffix("RETURNING id, code, description").
		ToSql()
	if err != nil {
		return domain.Role{}, fmt.Errorf("Build insert: %w", err)
	}

	var role domain.Role
	err = r.pool.QueryRow(ctx, query, args...).Scan(&role.ID, &role.Code, &role.Description)
	if err != nil {
		return domain.Role{}, fmt.Errorf("exec insert: %w", err)
	}
	return role, nil
}

func (r *PostgresRepo) GetByCode(ctx context.Context, code string) (domain.Role, error) {
	query, args, err := psql.
		Select("id", "code", "description").
		From("roles").
		Where("code = ?", code).
		ToSql()
	if err != nil {
		return domain.Role{}, fmt.Errorf("build query: %w", err)
	}

	var role domain.Role
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&role.ID,
		&role.Code,
		&role.Description,
	)
	if err != nil {
		return domain.Role{}, fmt.Errorf("role with code %s: %w", code, err)
	}
	return role, nil
}

func (r *PostgresRepo) List(ctx context.Context) ([]domain.Role, error) {
	query, args, err := psql.
		Select("id", "code", "description").
		From("roles").
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

	out := make([]domain.Role, 0)
	for rows.Next() {
		var roles domain.Role
		if err := rows.Scan(
			&roles.ID,
			&roles.Code,
			&roles.Description,
		); err != nil {
			return nil, fmt.Errorf("Scan: %w", err)
		}
		out = append(out, roles)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}
