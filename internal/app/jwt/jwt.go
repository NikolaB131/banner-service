package jwt

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	jwt.RegisteredClaims
	UserID   string `json:"id"`
	Username string `json:"username"`
}

func Generate(signSecret string, id string, username string) (string, error) {
	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{},
		UserID:           id,
		Username:         username,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(signSecret))
	return signedToken, err
}

func Parse(token string, signSecret string) (*JWTClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(signSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsedToken.Claims.(*JWTClaims)
	if !parsedToken.Valid || !ok {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
