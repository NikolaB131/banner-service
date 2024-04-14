package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/NikolaB131-org/banner-service/internal/entity"
	"github.com/NikolaB131-org/banner-service/internal/repository"
	redisPkg "github.com/NikolaB131-org/banner-service/pkg/redis"
	"github.com/redis/go-redis/v9"
)

type BannerRepository struct {
	Client    *redis.Client
	BannerTTL time.Duration
}

const (
	bannerKey     string = "banner:feature_id=%d,tag_id=%d"
	bannerDataKey string = "banner-data:%v"
)

func NewBannerRepository(client *redisPkg.Redis, bannerTTL time.Duration) *BannerRepository {
	return &BannerRepository{Client: client.Client, BannerTTL: bannerTTL}
}

func (r *BannerRepository) Banner(ctx context.Context, featureID int, tagID int) (entity.Banner, error) {
	bannerDataID, err := r.Client.Get(ctx, fmt.Sprintf(bannerKey, featureID, tagID)).Result()
	if errors.Is(err, redis.Nil) {
		return entity.Banner{}, repository.ErrNotFound
	}
	if err != nil {
		return entity.Banner{}, fmt.Errorf("redis get banner uuid failed: %w", err)
	}
	bannerData, err := r.Client.Get(ctx, fmt.Sprintf(bannerDataKey, bannerDataID)).Result()
	if err != nil {
		return entity.Banner{}, fmt.Errorf("redis get banner data failed: %w", err)
	}

	var banner entity.Banner

	if err := json.Unmarshal([]byte(bannerData), &banner); err != nil {
		return entity.Banner{}, fmt.Errorf("failed parsing string to json: %w", err)
	}

	return banner, nil
}

func (r *BannerRepository) SaveBanner(ctx context.Context, banner entity.Banner) error {
	_, err := r.Client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		bannerData, err := json.Marshal(banner)
		if err != nil {
			return fmt.Errorf("failed parsing banner to json string: %w", err)
		}

		for _, tagID := range banner.TagIDs {
			err := pipe.SetEx(ctx, fmt.Sprintf(bannerKey, banner.FeatureID, tagID), banner.ID, r.BannerTTL).Err()
			if err != nil {
				return fmt.Errorf("redis setex failed: %w", err)
			}
		}
		err = pipe.SetEx(ctx, fmt.Sprintf(bannerDataKey, banner.ID), bannerData, r.BannerTTL).Err()
		if err != nil {
			return fmt.Errorf("redis setex failed: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("redis pipeline failed: %w", err)
	}

	return nil
}
