package v1

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/NikolaB131-org/banner-service/internal/controller/http/v1/middlewares"
	"github.com/NikolaB131-org/banner-service/internal/service"
	"github.com/gin-gonic/gin"
)

type (
	BannerRoutes struct {
		bannerService service.BannerService
	}

	BannerGetQuery struct {
		FeatureID *int `form:"feature_id"`
		TagID     *int `form:"tag_id"`
		Limit     *int `form:"limit"`
		Offset    *int `form:"offset"`
	}

	BannerCreateBody struct {
		TagIDs    []int          `json:"tag_ids" binding:"required"`
		FeatureID *int           `json:"feature_id" binding:"required"`
		Content   map[string]any `json:"content" binding:"required"`
		IsActive  *bool          `json:"is_active" binding:"required"`
	}

	BannerUpdateBody struct {
		TagIDs    []int          `json:"tag_ids"`
		FeatureID *int           `json:"feature_id"`
		Content   map[string]any `json:"content"`
		IsActive  *bool          `json:"is_active"`
	}
)

func newBannerRoutes(g *gin.RouterGroup, middlewares middlewares.Middlewares, bannerService service.BannerService) {
	bannerR := BannerRoutes{bannerService: bannerService}

	banner := g.Group("/banner", middlewares.OnlyAuth(), middlewares.OnlyAdmin())
	{
		banner.GET("/", bannerR.get)
		banner.POST("/", bannerR.create)
		banner.PATCH("/:id", bannerR.update)
		banner.DELETE("/:id", bannerR.deleteById)
	}
}

func (r *BannerRoutes) get(c *gin.Context) {
	var query BannerGetQuery

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parsing error"})
		return
	}

	banners, err := r.bannerService.GetBanners(c, query.FeatureID, query.TagID, query.Limit, query.Offset)
	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed get banners"})
		return
	}

	c.JSON(http.StatusOK, banners)
}

func (r *BannerRoutes) create(c *gin.Context) {
	var body BannerCreateBody

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body parsing error"})
		return
	}
	if len(body.TagIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag_ids is required"})
		return
	}
	if len(body.Content) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
		return
	}

	id, err := r.bannerService.Create(c, body.TagIDs, *body.FeatureID, body.Content, *body.IsActive)
	if err != nil {
		slog.Error(err.Error())
		switch {
		case errors.Is(err, service.ErrBannerFeatureNotExists) || errors.Is(err, service.ErrBannerTagNotExists):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrBannerAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create banner"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"banner_id": id})
}

func (r *BannerRoutes) update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "specified id is not a number"})
		return
	}

	var body BannerUpdateBody

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body parsing error"})
		return
	}
	if body.TagIDs != nil && len(body.TagIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag_ids must not be empty"})
		return
	}
	if body.Content != nil && len(body.Content) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content must not be empty"})
		return
	}

	err = r.bannerService.Update(c, id, body.TagIDs, body.FeatureID, body.Content, body.IsActive)
	if err != nil {
		slog.Error(err.Error())
		switch {
		case errors.Is(err, service.ErrBannerNotFound):
			c.Status(http.StatusNotFound)
		case errors.Is(err, service.ErrBannerAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update banner"})
		}
		return
	}

	c.Status(http.StatusOK)
}

func (r *BannerRoutes) deleteById(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "specified id is not a number"})
		return
	}

	err = r.bannerService.DeleteByID(c, id)
	if err != nil {
		slog.Error(err.Error())
		switch {
		case errors.Is(err, service.ErrBannerNotFound):
			c.Status(http.StatusNotFound)
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete banner"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
