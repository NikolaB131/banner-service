package v1

import (
	"github.com/NikolaB131-org/banner-service/internal/controller/http/v1/middlewares"
	"github.com/NikolaB131-org/banner-service/internal/service"
	"github.com/gin-gonic/gin"
)

func NewRouter(r *gin.Engine, middlewares middlewares.Middlewares, authService service.AuthService, bannerService service.BannerService) {
	v1 := r.Group("/v1")
	{
		newAuthRoutes(v1, authService)
		newBannerRoutes(v1, middlewares, bannerService)
		newUserBannerRoutes(v1, middlewares, bannerService)
	}
}
