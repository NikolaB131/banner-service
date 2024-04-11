package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/NikolaB131-org/banner-service/internal/entity"
	"github.com/NikolaB131-org/banner-service/pkg/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	UserRepository struct {
		Pool *pgxpool.Pool
	}
)

var (
	ErrUserNotFound = errors.New("user not found")
)

func NewUserRepository(pg *postgres.Postgres) *UserRepository {
	return &UserRepository{Pool: pg.Pool}
}

func (r *UserRepository) SaveUser(ctx context.Context, user entity.User) (string, error) {
	var id string

	row := r.Pool.QueryRow(ctx, "INSERT INTO users (username, password_hash) VALUES($1, $2) RETURNING id", user.Username, user.PasswordHash)
	err := row.Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to scan db row: %w", err)
	}

	return id, nil
}

func (r *UserRepository) User(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User

	row := r.Pool.QueryRow(ctx, "SELECT id, username, password_hash, role FROM users WHERE username = $1", username)
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to scan db row: %w", err)
	}

	return &user, nil
}
