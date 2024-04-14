package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/NikolaB131-org/banner-service/config"
	postgresRepo "github.com/NikolaB131-org/banner-service/internal/repository/postgres"
	redisRepo "github.com/NikolaB131-org/banner-service/internal/repository/redis"
	"github.com/NikolaB131-org/banner-service/internal/service"
	"github.com/NikolaB131-org/banner-service/pkg/postgres"
	"github.com/NikolaB131-org/banner-service/pkg/redis"
	"github.com/stretchr/testify/suite"
)

type UserBannerSuite struct {
	suite.Suite
	BannerService  *service.Banner
	BaseUrl        string
	TestUserToken  string
	TestAdminToken string
}

func TestUserBannerSuite(t *testing.T) {
	suite.Run(t, new(UserBannerSuite))
}

func (suite *UserBannerSuite) SetupSuite() {
	ctx := context.Background()
	configPath := "/app/config.yml"
	config, err := config.NewConfig(&configPath)
	if err != nil {
		panic(err)
	}
	suite.BaseUrl = fmt.Sprintf("http://localhost:%d/v1/user_banner", config.HTTP.Port)
	pg, err := postgres.New(config.DB.Url)
	if err != nil {
		panic(err)
	}
	redisClient, err := redis.New(config.Redis.Url)
	if err != nil {
		panic(err)
	}
	userRepository := postgresRepo.NewUserRepository(pg)
	bannerRepository := postgresRepo.NewBannerRepository(pg)
	bannerCacheRepository := redisRepo.NewBannerRepository(redisClient, config.Redis.BannerTTL)
	tagRepository := postgresRepo.NewTagRepository(pg)
	featureRepository := postgresRepo.NewFeatureRepository(pg)
	authService := service.NewAuthService(userRepository, config.Auth.SignSecret, config.Auth.TokenTTL)
	bannerService := service.NewBannerService(bannerRepository, bannerCacheRepository, tagRepository, featureRepository)
	suite.BannerService = bannerService

	_, err = authService.RegisterUser(ctx, "testuser", "testpass")
	if err != nil {
		panic(err)
	}
	token, err := authService.Login(ctx, "testuser", "testpass")
	if err != nil {
		panic(err)
	}
	suite.TestUserToken = fmt.Sprintf("Bearer %s", token)
	token, err = authService.Login(ctx, "admin", "admin")
	if err != nil {
		panic(err)
	}
	suite.TestAdminToken = fmt.Sprintf("Bearer %s", token)

	_, err = bannerService.Create(ctx, []int{20, 21}, 10, map[string]any{"info": "123"}, true) // active banner
	if err != nil {
		panic(err)
	}
	_, err = bannerService.Create(ctx, []int{25}, 10, map[string]any{"memes_counter": 25}, false) // inactive banner
	if err != nil {
		panic(err)
	}
}

func (s *UserBannerSuite) TestUserBannerRoutes_GetBanner() {
	testCases := []struct {
		reqHeaders    map[string]string
		reqBody       string
		reqQuery      string
		resStatusCode int
		resBody       string
	}{
		{
			reqHeaders:    map[string]string{"Authorization": s.TestUserToken},
			reqQuery:      "?tag_id=20&feature_id=10",
			resStatusCode: http.StatusOK,
			resBody:       `{"info": "123"}`,
		},
		{
			reqHeaders:    map[string]string{"Authorization": s.TestUserToken},
			reqQuery:      "?tag_id=21&feature_id=10",
			resStatusCode: http.StatusOK,
			resBody:       `{"info": "123"}`,
		},
		{
			reqHeaders:    map[string]string{"Authorization": s.TestUserToken},
			reqQuery:      "?tag_id=2abc&feature_id=11",
			resStatusCode: http.StatusBadRequest,
			resBody:       `{"error": "query parsing error"}`,
		},
		{
			reqHeaders:    map[string]string{"Authorization": s.TestUserToken},
			reqQuery:      "?tag_id=20",
			resStatusCode: http.StatusBadRequest,
			resBody:       `{"error": "feature_id must be specified"}`,
		},
		{
			reqHeaders:    map[string]string{"Authorization": s.TestUserToken},
			reqQuery:      "?feature_id=10",
			resStatusCode: http.StatusBadRequest,
			resBody:       `{"error": "tag_id must be specified"}`,
		},
		{
			reqHeaders:    map[string]string{"Authorization": s.TestUserToken},
			reqQuery:      "?tag_id=20&feature_id=11",
			resStatusCode: http.StatusNotFound,
		},
		{
			reqQuery:      "?tag_id=20&feature_id=10",
			resStatusCode: http.StatusUnauthorized,
		},
		{
			reqHeaders:    map[string]string{"Authorization": s.TestUserToken},
			reqQuery:      "?tag_id=25&feature_id=10",
			resStatusCode: http.StatusForbidden,
		},
		{
			reqHeaders:    map[string]string{"Authorization": s.TestAdminToken},
			reqQuery:      "?tag_id=25&feature_id=10",
			resStatusCode: http.StatusOK,
			resBody:       `{"memes_counter": 25}`,
		},
	}

	for _, testCase := range testCases {
		req, _ := http.NewRequest(
			http.MethodGet,
			s.BaseUrl+testCase.reqQuery,
			strings.NewReader(testCase.reqBody),
		)
		for key, value := range testCase.reqHeaders {
			req.Header.Add(key, value)
		}

		res, _ := http.DefaultClient.Do(req)
		parsedBody, _ := io.ReadAll(res.Body)

		s.Equal(testCase.resStatusCode, res.StatusCode)
		if testCase.resBody != "" {
			s.JSONEq(testCase.resBody, string(parsedBody))
		}
	}
}

func (s *UserBannerSuite) TestUserBannerRoutes_GetBannerLastRevision() {
	bannerID, err := s.BannerService.Create(context.Background(), []int{27}, 15, map[string]any{"company": "Avito"}, true)
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest(http.MethodGet, s.BaseUrl+"?tag_id=27&feature_id=15", nil)
	req.Header.Add("Authorization", s.TestUserToken)
	res, _ := http.DefaultClient.Do(req)
	parsedBody, _ := io.ReadAll(res.Body)
	var resBody1 struct {
		Token string `json:"token"`
	}
	json.Unmarshal(parsedBody, &resBody1)

	s.Equal(http.StatusOK, res.StatusCode)
	s.JSONEq(`{"company": "Avito"}`, string(parsedBody))

	err = s.BannerService.Update(context.Background(), bannerID, nil, nil, map[string]any{"job": "Avito"}, nil)
	if err != nil {
		panic(err)
	}

	req, _ = http.NewRequest(http.MethodGet, s.BaseUrl+"?tag_id=27&feature_id=15", nil)
	req.Header.Add("Authorization", s.TestUserToken)
	res, _ = http.DefaultClient.Do(req)
	parsedBody, _ = io.ReadAll(res.Body)
	var resBody2 struct {
		Token string `json:"token"`
	}
	json.Unmarshal(parsedBody, &resBody2)

	s.Equal(http.StatusOK, res.StatusCode)
	s.JSONEq(`{"company": "Avito"}`, string(parsedBody))

	req, _ = http.NewRequest(http.MethodGet, s.BaseUrl+"?tag_id=27&feature_id=15&use_last_revision=true", nil)
	req.Header.Add("Authorization", s.TestUserToken)
	res, _ = http.DefaultClient.Do(req)
	parsedBody, _ = io.ReadAll(res.Body)
	var resBody3 struct {
		Token string `json:"token"`
	}
	json.Unmarshal(parsedBody, &resBody3)

	s.Equal(http.StatusOK, res.StatusCode)
	s.JSONEq(`{"job": "Avito"}`, string(parsedBody))
}
