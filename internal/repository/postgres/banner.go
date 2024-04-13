package postgres

import (
	"context"
	"fmt"

	"github.com/NikolaB131-org/banner-service/internal/entity"
	"github.com/NikolaB131-org/banner-service/internal/repository"
	"github.com/NikolaB131-org/banner-service/pkg/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BannerRepository struct {
	Pool *pgxpool.Pool
}

func NewBannerRepository(pg *postgres.Postgres) *BannerRepository {
	return &BannerRepository{Pool: pg.Pool}
}

func (r *BannerRepository) IsExistsById(ctx context.Context, id int) (bool, error) {
	isExists := false
	err := r.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM banners WHERE id = $1)", id).Scan(&isExists)
	if err != nil {
		return false, fmt.Errorf("failed query: %w", err)
	}

	return isExists, nil
}

func (r *BannerRepository) IsExists(ctx context.Context, featureID int, tagID int) (bool, error) {
	isExists := false
	err := r.Pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM banners INNER JOIN banner_tags ON id = banner_id WHERE banners.feature_id = $1 AND banner_tags.tag_id = $2)",
		featureID, tagID,
	).Scan(&isExists)
	if err != nil {
		return false, fmt.Errorf("failed query: %w", err)
	}

	return isExists, nil
}

func (r *BannerRepository) Banners(ctx context.Context, featureID *int, tagID *int, limit *int, offset *int) ([]entity.Banner, error) {
	queryWherePart := ""
	if featureID != nil || tagID != nil {
		queryWherePart = "WHERE "
		if featureID != nil {
			queryWherePart += "feature_id = @featureID"
			if tagID != nil {
				queryWherePart += " AND "
			}
		}
		if tagID != nil {
			queryWherePart += "EXISTS (SELECT 1 FROM banner_tags WHERE id = banner_id AND tag_id = @tagID)"
		}
	}
	queryLimitPart := ""
	if limit != nil {
		queryLimitPart = "LIMIT @limit"
	}

	query := fmt.Sprintf(`
SELECT
	id,
	ARRAY(SELECT tag_id FROM banner_tags WHERE banner_id = id) AS tag_ids,
	feature_id,
	content,
	is_active,
	created_at,
	updated_at
FROM banners
%s
OFFSET @offset
%s`, queryWherePart, queryLimitPart)

	rows, err := r.Pool.Query(ctx, query,
		pgx.NamedArgs{
			"featureID": featureID,
			"tagID":     tagID,
			"limit":     limit,
			"offset":    offset,
		},
	)
	if err != nil {
		return []entity.Banner{}, fmt.Errorf("failed query: %w", err)
	}
	banners, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.Banner])
	if err != nil {
		return []entity.Banner{}, fmt.Errorf("failed collecting rows: %w", err)
	}

	return banners, nil
}

func (r *BannerRepository) BannerById(ctx context.Context, id int) (entity.Banner, error) {
	var banner entity.Banner
	err := r.Pool.QueryRow(ctx, `
SELECT
	id,
	ARRAY(SELECT tag_id FROM banner_tags WHERE banner_id = id) AS tag_ids,
	feature_id,
	content,
	is_active,
	created_at,
	updated_at
FROM banners WHERE id = $1`, id).Scan(&banner)
	if err != nil {
		return entity.Banner{}, fmt.Errorf("failed query: %w", err)
	}

	return banner, nil
}

func (r *BannerRepository) SaveBanner(ctx context.Context, tagIDs []int, featureID int, content map[string]any, isActive bool) (int, error) {
	var bannerID int

	err := r.Pool.QueryRow(ctx,
		"INSERT INTO banners (feature_id, content, is_active) VALUES($1, $2, $3) RETURNING id",
		featureID, content, isActive,
	).Scan(&bannerID)
	if err != nil {
		return 0, fmt.Errorf("failed query: %w", err)
	}

	var rows [][]any
	for _, tagID := range tagIDs {
		rows = append(rows, []any{bannerID, tagID})
	}
	_, err = r.Pool.CopyFrom(ctx, pgx.Identifier{"banner_tags"}, []string{"banner_id", "tag_id"}, pgx.CopyFromRows(rows))
	if err != nil {
		return 0, fmt.Errorf("failed to insert ids to banner_tags: %w", err)
	}

	return bannerID, nil
}

func (r *BannerRepository) UpdateBanner(ctx context.Context, bannerID int, tagIDs []int, featureID *int, content map[string]any, isActive *bool) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	update := func(dbField string, value any) error {
		query := fmt.Sprintf("UPDATE banners SET %s = $1 WHERE id = $2", dbField)
		_, err := tx.Exec(ctx, query, value, bannerID)
		if err != nil {
			return fmt.Errorf("failed to update %s: %w", dbField, err)
		}
		return nil
	}

	if featureID != nil {
		err := update("feature_id", *featureID)
		if err != nil {
			return err
		}
	}
	if content != nil {
		err := update("content", content)
		if err != nil {
			return err
		}
	}
	if isActive != nil {
		err := update("is_active", *isActive)
		if err != nil {
			return err
		}
	}
	if tagIDs != nil {
		_, err = tx.Exec(ctx, "DELETE FROM banner_tags WHERE banner_id = $1", bannerID)
		if err != nil {
			return fmt.Errorf("failed to delete banner tags: %w", err)
		}

		var oldBanner entity.Banner
		if featureID == nil {
			oldBanner, err = r.BannerById(ctx, bannerID)
			if err != nil {
				return fmt.Errorf("failed to find old banner: %w", err)
			}
		}

		var rows [][]any
		for _, tagID := range tagIDs {
			var err error
			var isExists bool
			if featureID == nil {
				isExists, err = r.IsExists(ctx, oldBanner.FeatureID, tagID)
			} else {
				isExists, err = r.IsExists(ctx, *featureID, tagID)
			}
			if err != nil {
				return fmt.Errorf("failed to check new banner conflicts with old: %w", err)
			}
			if isExists {
				return repository.ErrAlreadyExists
			}
			rows = append(rows, []any{bannerID, tagID})
		}
		_, err = tx.CopyFrom(ctx, pgx.Identifier{"banner_tags"}, []string{"banner_id", "tag_id"}, pgx.CopyFromRows(rows))
		if err != nil {
			return fmt.Errorf("failed to insert ids to banner_tags: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *BannerRepository) DeleteBannerByID(ctx context.Context, id int) error {
	res, err := r.Pool.Exec(ctx, "DELETE FROM banners WHERE id = $1", id)
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to delete banner: %w", err)
	}

	return nil
}
