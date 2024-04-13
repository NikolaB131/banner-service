package postgres

import (
	"context"
	"fmt"

	"github.com/NikolaB131-org/banner-service/pkg/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TagRepository struct {
	Pool *pgxpool.Pool
}

func NewTagRepository(pg *postgres.Postgres) *TagRepository {
	return &TagRepository{Pool: pg.Pool}
}

func (r *TagRepository) IsExist(ctx context.Context, id int) (bool, error) {
	isExists := false
	err := r.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM tags WHERE id = $1)", id).Scan(&isExists)
	if err != nil {
		return false, fmt.Errorf("failed to scan db row: %w", err)
	}

	return isExists, nil
}
