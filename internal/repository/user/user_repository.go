package userrepo

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

func (r *PostgresRepo) Insert(ctx context.Context, u domain.User) (domain.User, error) {
	query, args, err := psql.
		Insert("users").
		Columns("external_id", "username", "password_hash", "email").
		Values(u.ExternalID, u.Username, u.PasswordHash, u.Email).
		Suffix("RETURNING id, external_id, username, password_hash, email, created_at").
		ToSql()
	if err != nil {
		return domain.User{}, fmt.Errorf("Build insert: %w", err)
	}

	var user domain.User
	err = r.pool.QueryRow(ctx, query, args...).Scan(&user.ID, &user.ExternalID, &user.Username, &user.PasswordHash, &user.Email, &user.CreatedAt)
	if err != nil {
		return domain.User{}, fmt.Errorf("exec insert: %w", err)
	}
	return user, nil
}

func (r *PostgresRepo) GetByUsername(ctx context.Context, username string) (domain.User, error) {
	query, args, err := psql.
		Select("id", "external_id", "username", "password_hash", "email", "created_at").
		From("users").
		Where("username = ?", username).
		ToSql()
	if err != nil {
		return domain.User{}, fmt.Errorf("build query: %w", err)
	}

	var user domain.User
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.ExternalID,
		&user.Username,
		&user.PasswordHash,
		&user.Email,
		&user.CreatedAt,
	)
	if err != nil {
		return domain.User{}, fmt.Errorf("user %s not found: %w", username, err)
	}
	return user, nil
}
