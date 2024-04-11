package main

import (
	"context"
	"fmt"

	"github.com/NikolaB131-org/banner-service/config"
	"github.com/NikolaB131-org/banner-service/internal/app"
	v1 "github.com/NikolaB131-org/banner-service/internal/controller/http/v1"
	postgresRepo "github.com/NikolaB131-org/banner-service/internal/repository/postgres"
	"github.com/NikolaB131-org/banner-service/internal/service/auth"
	"github.com/NikolaB131-org/banner-service/pkg/postgres"
	"github.com/gin-gonic/gin"
)

func main() {
	// Config
	config, err := config.NewConfig()
	if err != nil {
		panic(err.Error())
	}

	// Logger
	app.InitLogger(config.Logger.Level)

	// Postgres db
	pg, err := postgres.New(config.DB.Url)
	if err != nil {
		panic(err.Error())
	}
	defer pg.Close()

	// Repositories init
	userRepository := postgresRepo.NewUserRepository(pg)

	// Services
	authService := auth.New(userRepository, config.Auth.SignSecret, config.Auth.TokenTTL)

	// Creating admin user
	adminUser, _ := userRepository.User(context.Background(), config.Auth.AdminUsername)
	if adminUser == nil {
		adminID, err := authService.RegisterUser(context.Background(), config.Auth.AdminUsername, config.Auth.AdminPassword)
		if err != nil && err != auth.ErrUserAlreadyExists {
			panic(fmt.Sprintf("unable to create admin user: %s", err.Error()))
		}
		err = authService.MakeAdmin(context.Background(), adminID)
		if err != nil {
			panic(fmt.Sprintf("unable to grant permissions to admin user: %s", err.Error()))
		}
	}

	// Routes
	r := gin.New()
	v1.NewRouter(r, config, authService)

	r.Run(fmt.Sprintf(":%d", config.HTTP.Port))
}
