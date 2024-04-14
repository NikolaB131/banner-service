package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/NikolaB131-org/banner-service/config"
	"github.com/NikolaB131-org/banner-service/internal/app"
	v1 "github.com/NikolaB131-org/banner-service/internal/controller/http/v1"
	"github.com/NikolaB131-org/banner-service/internal/controller/http/v1/middlewares"
	postgresRepo "github.com/NikolaB131-org/banner-service/internal/repository/postgres"
	redisRepo "github.com/NikolaB131-org/banner-service/internal/repository/redis"
	"github.com/NikolaB131-org/banner-service/internal/service"
	"github.com/NikolaB131-org/banner-service/pkg/postgres"
	"github.com/NikolaB131-org/banner-service/pkg/redis"
	"github.com/gin-gonic/gin"
)

func main() {
	// Config
	config, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	// Logger
	app.InitLogger(config.Logger.Level)

	// Postgres
	pg, err := postgres.New(config.DB.Url)
	if err != nil {
		panic(err)
	}
	defer pg.Close()

	// Redis
	redisClient, err := redis.New(config.Redis.Url)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := redisClient.Close()
		if err != nil {
			panic(err)
		}
	}()

	// Repositories init
	userRepository := postgresRepo.NewUserRepository(pg)
	bannerRepository := postgresRepo.NewBannerRepository(pg)
	bannerCacheRepository := redisRepo.NewBannerRepository(redisClient, config.Redis.BannerTTL)
	tagRepository := postgresRepo.NewTagRepository(pg)
	featureRepository := postgresRepo.NewFeatureRepository(pg)

	// Services
	authService := service.NewAuthService(userRepository, config.Auth.SignSecret, config.Auth.TokenTTL)
	bannerService := service.NewBannerService(bannerRepository, bannerCacheRepository, tagRepository, featureRepository)

	// Creating admin user
	adminID, err := authService.RegisterUser(context.Background(), config.Auth.AdminUsername, config.Auth.AdminPassword)
	if !errors.Is(err, service.ErrUserAlreadyExists) {
		if err != nil {
			panic(fmt.Sprintf("unable to create admin user: %s", err.Error()))
		}
		err = authService.MakeAdmin(context.Background(), adminID)
		if err != nil {
			panic(fmt.Sprintf("unable to grant permissions to admin user: %s", err.Error()))
		}
	}

	// Middlewares
	middlewares := middlewares.New(config, userRepository)

	// Routes
	r := gin.New()
	v1.NewRouter(r, middlewares, authService, bannerService)

	r.Run(fmt.Sprintf(":%d", config.HTTP.Port))
}
