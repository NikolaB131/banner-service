package repository

import (
	"context"
	"errors"

	"github.com/NikolaB131-org/banner-service/internal/entity"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type User interface {
	SaveUser(ctx context.Context, user entity.User) (string, error)
	User(ctx context.Context, username string) (entity.User, error)
	GrantAdminPermission(ctx context.Context, userID string) error
}

type Banner interface {
	IsExistsById(ctx context.Context, id int) (bool, error)
	IsExists(ctx context.Context, featureID int, tagID int) (bool, error)
	Banners(ctx context.Context, featureID *int, tagID *int, limit *int, offset *int) ([]entity.Banner, error)
	BannerById(ctx context.Context, id int) (entity.Banner, error)
	SaveBanner(ctx context.Context, tagIDs []int, featureID int, content map[string]any, isActive bool) (int, error)
	UpdateBanner(ctx context.Context, bannerID int, tagIDs []int, featureID *int, content map[string]any, isActive *bool) error
	DeleteBannerByID(ctx context.Context, id int) error
}

type Feature interface {
	IsExist(ctx context.Context, id int) (bool, error)
}

type Tag interface {
	IsExist(ctx context.Context, id int) (bool, error)
}
