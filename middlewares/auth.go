package middlewares

import (
	"gocash/config"
	"gocash/models"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Authentication middleware
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := &models.Claims{}
		var token string

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "token_required",
				"message": "Auth token is required",
			})
			return
		}
		splitToken := strings.Split(authHeader, "Bearer ")
		if len(splitToken) > 1 {
			token = splitToken[1]
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "token_wrong",
				"message": "Invalid token",
			})
			return
		}
		tkn, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
			return config.GlobalConfig.JWT_SECRET, nil
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   err.Error(),
				"message": "Couldn't parse token",
			})
			return
		}

		if claims.ExpiresAt.Unix() < time.Now().Local().Unix() {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "token_expired",
				"message": "Token expired",
			})
			return
		}

		if !tkn.Valid {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "invalid_token",
				"message": "Invalid token",
			})
			return
		}
		c.Next()
	}
}

// GenerateJWT creates access and refresh tokens with user's username
func GenerateJWT(username string) (token models.Tokens, err error) {
	// Create access token
	os.Getenv("ACCESS_TOKEN_TIMEOUT")
	accessTokenExp := time.Now().Add(time.Duration(config.GlobalConfig.ACCESS_TOKEN_TIMEOUT) * time.Second)
	accessClaims := &models.Claims{
		User: models.User{
			Username: username,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: accessTokenExp},
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	token.AccessToken, err = accessToken.SignedString(config.GlobalConfig.JWT_SECRET)
	if err != nil {
		return models.Tokens{}, err
	}

	// Create refresh token
	refreshTokenExp := time.Now().Add(time.Duration(config.GlobalConfig.REFRESH_TOKEN_TIMEOUT) * time.Minute)
	refreshClaims := &models.Claims{
		User: models.User{
			Username: username,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: refreshTokenExp},
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	token.RefreshToken, err = refreshToken.SignedString(config.GlobalConfig.JWT_SECRET)
	if err != nil {
		return models.Tokens{}, err
	}

	return token, nil
}

func RefreshToken(claims *models.Claims) (token models.Tokens, err error) {
	expirationTime := time.Now().Add(time.Duration(config.GlobalConfig.ACCESS_TOKEN_TIMEOUT) * time.Second)

	claims.ExpiresAt = &jwt.NumericDate{Time: expirationTime}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.AccessToken, err = accessToken.SignedString(config.GlobalConfig.JWT_SECRET)
	if err != nil {
		return models.Tokens{}, err
	}

	expirationTime = time.Now().Add(time.Duration(config.GlobalConfig.REFRESH_TOKEN_TIMEOUT) * time.Second)

	claims.ExpiresAt = &jwt.NumericDate{Time: expirationTime}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	token.RefreshToken, err = refreshToken.SignedString(config.GlobalConfig.JWT_SECRET)
	if err != nil {
		return models.Tokens{}, err
	}

	return token, nil
}
