package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/NikolaB131-org/banner-service/internal/app/jwt"
	"github.com/NikolaB131-org/banner-service/internal/entity"
	"github.com/NikolaB131-org/banner-service/internal/repository"
	"github.com/NikolaB131-org/banner-service/internal/repository/postgres"
	"golang.org/x/crypto/bcrypt"
)

type (
	AuthService interface {
		Login(ctx context.Context, username string, password string) (string, error)
		RegisterUser(ctx context.Context, username string, password string) (string, error)
		MakeAdmin(ctx context.Context, userID string) error
	}

	Auth struct {
		userRepository repository.User
		signSecret     string
		tokenTTL       time.Duration
	}
)

var (
	ErrUserAlreadyExists  = errors.New("user with this username already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func New(userRepository repository.User, signSecret string, tokenTTL time.Duration) *Auth {
	return &Auth{
		userRepository: userRepository,
		signSecret:     signSecret,
		tokenTTL:       tokenTTL,
	}
}

func (a *Auth) Login(ctx context.Context, username string, password string) (string, error) {
	user, err := a.userRepository.User(ctx, username)
	if err != nil {
		return "", fmt.Errorf("failed to check if user exists: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	signedToken, err := jwt.Generate(a.signSecret, a.tokenTTL, user.ID, user.Username)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func (a *Auth) RegisterUser(ctx context.Context, username string, password string) (string, error) {
	user, err := a.userRepository.User(ctx, username)
	if err != nil && !errors.Is(err, postgres.ErrUserNotFound) {
		return "", fmt.Errorf("failed to check if user exists: %w", err)
	}
	if user != nil {
		return "", ErrUserAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate password hash: %w", err)
	}

	userId, err := a.userRepository.SaveUser(ctx, entity.User{Username: username, PasswordHash: passwordHash})
	if err != nil {
		return "", fmt.Errorf("failed to save user: %w", err)
	}

	slog.Info(fmt.Sprintf("registering user with username: %s", username))

	return userId, nil
}

func (a *Auth) MakeAdmin(ctx context.Context, userID string) error {
	return a.userRepository.GrantAdminPermission(ctx, userID)
}
