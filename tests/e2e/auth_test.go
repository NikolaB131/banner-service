package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/NikolaB131-org/banner-service/config"
	"github.com/stretchr/testify/suite"
)

type AuthSuite struct {
	suite.Suite
	BaseUrl string
}

func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(AuthSuite))
}

func (suite *AuthSuite) SetupSuite() {
	configPath := "/app/config.yml"
	config, err := config.NewConfig(&configPath)
	if err != nil {
		panic(err)
	}
	suite.BaseUrl = fmt.Sprintf("http://localhost:%d/v1/auth", config.HTTP.Port)
}

func (s *AuthSuite) TestAuthRoutes_LoginAdmin() {
	body := `{"username": "admin", "password": "admin"}`

	res, _ := http.Post(s.BaseUrl+"/login", "application/json", strings.NewReader(body))
	parsedBody, _ := io.ReadAll(res.Body)
	var resBody struct {
		Token string `json:"token"`
	}
	json.Unmarshal(parsedBody, &resBody)

	s.Equal(http.StatusOK, res.StatusCode)
	s.NotEmpty(resBody.Token)
}

func (s *AuthSuite) TestAuthRoutes_RegisterLoginUser() {
	body := `{"username": "testuser2", "password": "qwerty"}`

	res, _ := http.Post(s.BaseUrl+"/register", "application/json", strings.NewReader(body))
	parsedBody, _ := io.ReadAll(res.Body)
	var resBodyRegister struct {
		UserID string `json:"user_id"`
	}
	json.Unmarshal(parsedBody, &resBodyRegister)
	s.Equal(http.StatusOK, res.StatusCode)
	s.NotEmpty(resBodyRegister.UserID)

	res, _ = http.Post(s.BaseUrl+"/login", "application/json", strings.NewReader(body))
	parsedBody, _ = io.ReadAll(res.Body)
	var resBodyLogin struct {
		Token string `json:"token"`
	}
	json.Unmarshal(parsedBody, &resBodyLogin)
	s.Equal(http.StatusOK, res.StatusCode)
	s.NotEmpty(resBodyLogin.Token)
}
