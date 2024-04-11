package v1

import (
	"log/slog"
	"net/http"

	"github.com/NikolaB131-org/banner-service/internal/service/auth"
	"github.com/gin-gonic/gin"
)

type AuthRoutes struct {
	authService auth.AuthService
}

type AuthBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func newAuthRoutes(g *gin.RouterGroup, authService auth.AuthService) {
	authR := AuthRoutes{authService: authService}

	auth := g.Group("/auth")
	{
		auth.POST("/login", authR.login)
		auth.POST("/register", authR.register)
	}
}

func (r *AuthRoutes) login(c *gin.Context) {
	var body AuthBody

	if err := c.BindJSON(&body); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if body.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}
	if body.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password is required"})
		return
	}

	token, err := r.authService.Login(c, body.Username, body.Password)
	if err != nil {
		slog.Error(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (r *AuthRoutes) register(c *gin.Context) {
	var body AuthBody

	if err := c.BindJSON(&body); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if body.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}
	if body.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password is required"})
		return
	}

	id, err := r.authService.RegisterUser(c, body.Username, body.Password)
	if err != nil {
		slog.Error(err.Error())
		switch err {
		case auth.ErrUserAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_id": id})
}
