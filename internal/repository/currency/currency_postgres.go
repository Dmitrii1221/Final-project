package currencyrepo

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

func (r *PostgresRepo) Insert(ctx context.Context, c domain.Currency) (domain.Currency, error) {
	query, args, err := psql.
		Insert("currencies").
		Columns("code", "display_name", "precision").
		Values(c.Code, c.DisplayName, c.Precision).
		Suffix("RETURNING id, code, display_name, precision, created_at").
		ToSql()
	if err != nil {
		return domain.Currency{}, fmt.Errorf("Build insert: %w", err)
	}

	var curr domain.Currency
	err = r.pool.QueryRow(ctx, query, args...).Scan(&curr.ID, &curr.Code, &curr.DisplayName, &curr.Precision, &curr.CreatedAt)
	if err != nil {
		return domain.Currency{}, fmt.Errorf("exec insert: %w", err)
	}
	return curr, nil
}

func (r *PostgresRepo) GetByCode(ctx context.Context, code string) (domain.Currency, error) {
	query, args, err := psql.
		Select("id", "code", "display_name", "precision", "created_at").
		From("currencies").
		Where("code = ?", code).
		ToSql()
	if err != nil {
		return domain.Currency{}, fmt.Errorf("build query: %w", err)
	}

	var curr domain.Currency
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&curr.ID,
		&curr.Code,
		&curr.DisplayName,
		&curr.Precision,
		&curr.CreatedAt,
	)
	if err != nil {
		return domain.Currency{}, fmt.Errorf("currency with code %s: %w", code, err)
	}
	return curr, nil
}

func (r *PostgresRepo) List(ctx context.Context) ([]domain.Currency, error) {
	query, args, err := psql.
		Select("id", "code", "display_name", "precision", "created_at").
		From("currencies").
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

	out := make([]domain.Currency, 0)
	for rows.Next() {
		var curr domain.Currency
		if err := rows.Scan(&curr.ID,
			&curr.Code,
			&curr.DisplayName,
			&curr.Precision,
			&curr.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("Scan: %w", err)
		}
		out = append(out, curr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}
