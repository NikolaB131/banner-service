package v1

import (
	"github.com/NikolaB131-org/banner-service/config"
	"github.com/NikolaB131-org/banner-service/internal/service/auth"
	"github.com/gin-gonic/gin"
)

func NewRouter(r *gin.Engine, config *config.Config, authService auth.AuthService) {
	v1 := r.Group("/v1")
	{
		newAuthRoutes(v1, authService)
	}
}
