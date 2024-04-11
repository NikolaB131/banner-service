package repository

import (
	"context"

	"github.com/NikolaB131-org/banner-service/internal/entity"
)

type User interface {
	SaveUser(ctx context.Context, user entity.User) (string, error)
	User(ctx context.Context, username string) (*entity.User, error)
}
