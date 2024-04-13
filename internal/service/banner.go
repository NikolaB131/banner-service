package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/NikolaB131-org/banner-service/internal/entity"
	"github.com/NikolaB131-org/banner-service/internal/repository"
)

type (
	BannerService interface {
		Get(ctx context.Context, featureID *int, tagID *int, limit *int, offset *int) ([]entity.Banner, error)
		Create(ctx context.Context, tagIDs []int, featureID int, content map[string]any, isActive bool) (int, error)
		Update(ctx context.Context, id int, tagIDs []int, featureID *int, content map[string]any, isActive *bool) error
		DeleteByID(ctx context.Context, id int) error
	}

	Banner struct {
		bannerRepository  repository.Banner
		tagRepository     repository.Tag
		featureRepository repository.Feature
	}
)

var (
	ErrBannerNotFound         = errors.New("banner not found")
	ErrBannerAlreadyExists    = errors.New("banner already exists")
	ErrBannerTagNotExists     = errors.New("banner tag not exists")
	ErrBannerFeatureNotExists = errors.New("banner feature not exists")
)

func NewBannerService(bannerRepository repository.Banner, tagRepository repository.Tag, featureRepository repository.Feature) *Banner {
	return &Banner{
		bannerRepository:  bannerRepository,
		tagRepository:     tagRepository,
		featureRepository: featureRepository,
	}
}

func (b *Banner) Get(ctx context.Context, featureID *int, tagID *int, limit *int, offset *int) ([]entity.Banner, error) {
	banners, err := b.bannerRepository.Banners(ctx, featureID, tagID, limit, offset)
	if err != nil {
		return []entity.Banner{}, fmt.Errorf("failed to get banners: %w", err)
	}

	return banners, nil
}

func (b *Banner) Create(ctx context.Context, tagIDs []int, featureID int, content map[string]any, isActive bool) (int, error) {
	IsFeatureExists, err := b.featureRepository.IsExist(ctx, featureID)
	if err != nil {
		return 0, fmt.Errorf("failed to check is feature exists: %w", err)
	}
	if !IsFeatureExists {
		return 0, ErrBannerFeatureNotExists
	}

	for _, tagID := range tagIDs {
		IsTagExists, err := b.tagRepository.IsExist(ctx, tagID)
		if err != nil {
			return 0, fmt.Errorf("failed to check is tag exists: %w", err)
		}
		if !IsTagExists {
			return 0, ErrBannerTagNotExists
		}

		isExists, err := b.bannerRepository.IsExists(ctx, featureID, tagID)
		if err != nil {
			return 0, fmt.Errorf("failed to check creating banner conflicts: %w", err)
		}
		if isExists {
			return 0, ErrBannerAlreadyExists
		}
	}
	id, err := b.bannerRepository.SaveBanner(ctx, tagIDs, featureID, content, isActive)
	if err != nil {
		return 0, fmt.Errorf("failed to create banner: %w", err)
	}

	return id, nil
}

func (b *Banner) Update(ctx context.Context, bannerID int, tagIDs []int, featureID *int, content map[string]any, isActive *bool) error {
	exists, err := b.bannerRepository.IsExistsById(ctx, bannerID)
	if err != nil {
		return fmt.Errorf("failed to check is banner exists: %w", err)
	}
	if !exists {
		return ErrBannerNotFound
	}

	err = b.bannerRepository.UpdateBanner(ctx, bannerID, tagIDs, featureID, content, isActive)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrAlreadyExists):
			return ErrBannerAlreadyExists
		default:
			return fmt.Errorf("failed to update banner: %w", err)
		}
	}

	return nil
}

func (b *Banner) DeleteByID(ctx context.Context, id int) error {
	err := b.bannerRepository.DeleteBannerByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return ErrBannerNotFound
		default:
			return fmt.Errorf("failed to delete banner: %w", err)
		}
	}

	return nil
}
