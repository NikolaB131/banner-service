package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/NikolaB131-org/banner-service/config"
	"github.com/NikolaB131-org/banner-service/internal/app/jwt"
	"github.com/gin-gonic/gin"
)

var (
	ErrParsingJWT = "error while parsing JWT token"
)

func OnlyAuth(config *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader("Authorization")
		if authorizationHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization token must be specified in headers"})
			return
		}

		splitToken := strings.Split(authorizationHeader, "Bearer ")
		token := splitToken[1]

		if len(token) < 2 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, err := jwt.Parse(token, config.Auth.SignSecret)
		if err != nil {
			slog.Warn(ErrParsingJWT)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

func OnlyAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("middleware 2")
		// role := "admin"
		c.Next()
	}
}
