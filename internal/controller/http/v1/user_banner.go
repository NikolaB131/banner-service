package v1

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/NikolaB131-org/banner-service/internal/controller/http/v1/middlewares"
	"github.com/NikolaB131-org/banner-service/internal/entity"
	"github.com/NikolaB131-org/banner-service/internal/service"
	"github.com/gin-gonic/gin"
)

type (
	UserBannerRoutes struct {
		bannerService service.BannerService
	}

	UserBannerGetQuery struct {
		FeatureID       *int  `form:"feature_id"`
		TagID           *int  `form:"tag_id"`
		UseLastRevision *bool `form:"use_last_revision"`
	}
)

func newUserBannerRoutes(g *gin.RouterGroup, middlewares middlewares.Middlewares, bannerService service.BannerService) {
	userBannerR := UserBannerRoutes{bannerService: bannerService}

	userBanner := g.Group("/user_banner", middlewares.OnlyAuth())
	{
		userBanner.GET("/", userBannerR.get)
	}
}

func (r *UserBannerRoutes) get(c *gin.Context) {
	var query UserBannerGetQuery

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parsing error"})
		return
	}
	if query.FeatureID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "feature_id must be specified"})
		return
	}
	if query.TagID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag_id must be specified"})
		return
	}

	var banner entity.Banner
	var err error
	if query.UseLastRevision != nil {
		banner, err = r.bannerService.GetBanner(c, *query.FeatureID, *query.TagID, *query.UseLastRevision)
	} else {
		banner, err = r.bannerService.GetBanner(c, *query.FeatureID, *query.TagID, false)
	}
	if err != nil {
		slog.Error(err.Error())
		switch {
		case errors.Is(err, service.ErrBannerNotFound):
			c.Status(http.StatusNotFound)
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get banner"})
		}
		return
	}

	userRole, _ := c.Get("user_role")
	if userRole == "user" && !banner.IsActive {
		c.Status(http.StatusForbidden)
		return
	}

	c.JSON(http.StatusOK, banner.Content)
}
