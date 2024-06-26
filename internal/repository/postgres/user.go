package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/NikolaB131-org/banner-service/internal/entity"
	"github.com/NikolaB131-org/banner-service/internal/repository"
	"github.com/NikolaB131-org/banner-service/pkg/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	Pool *pgxpool.Pool
}

func NewUserRepository(pg *postgres.Postgres) *UserRepository {
	return &UserRepository{Pool: pg.Pool}
}

func (r *UserRepository) SaveUser(ctx context.Context, user entity.User) (string, error) {
	var id string

	row := r.Pool.QueryRow(ctx,
		"INSERT INTO users (username, password_hash) VALUES($1, $2) RETURNING id",
		user.Username, user.PasswordHash,
	)
	err := row.Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to scan db row: %w", err)
	}

	return id, nil
}

func (r *UserRepository) User(ctx context.Context, username string) (entity.User, error) {
	var user entity.User

	row := r.Pool.QueryRow(ctx, "SELECT id, username, password_hash, role FROM users WHERE username = $1", username)
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role)

	if errors.Is(err, pgx.ErrNoRows) {
		return entity.User{}, repository.ErrNotFound
	}
	if err != nil {
		return entity.User{}, fmt.Errorf("failed to scan db row: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GrantAdminPermission(ctx context.Context, userID string) error {
	_, err := r.Pool.Exec(ctx, "UPDATE users SET role = 'admin' WHERE id = $1", userID)
	return err
}
