package controllers

import (
	"context"
	"gocash/config"
	"gocash/middlewares"
	"gocash/models"
	"gocash/pkg/db"
	"gocash/pkg/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func Login(ctx *gin.Context) {
	// Get body from the request
	var user models.User
	if err := ctx.BindJSON(&user); err != nil {
		logger.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "Couln't parse the request to user",
		})
		return
	}

	// Find the user with given data from database
	var dUser models.User
	err := db.DB.QueryRow(context.Background(), "select username, password from users where username = $1", user.Username).Scan(&dUser.Username, &dUser.Password)
	if err != nil {
		logger.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "Couldn't find user",
		})
		return
	}

	// Compare passwords
	if err := bcrypt.CompareHashAndPassword([]byte(dUser.Password), []byte(user.Password)); err != nil {
		logger.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "wrong_password",
			"message": "Invalid password",
		})
		return
	}

	// Generate new token
	tokens, err := middlewares.GenerateJWT(user.Username)
	if err != nil {
		logger.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "Coulnd't create token",
		})
		return
	}

	// Send success response
	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

func RefreshToken(ctx *gin.Context) {
	claims := &models.Claims{}

	// Get refresh token from request body
	token := models.Tokens{}
	if err := ctx.BindJSON(&token); err != nil {
		logger.Errorf("couldn't bind token body %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "Couldn't parse the request body",
		})
		return
	}

	// Validate jwt token
	tkn, err := jwt.ParseWithClaims(token.RefreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		return config.GlobalConfig.JWT_SECRET, nil
	})
	if err != nil {
		logger.Errorf("token didn't parse %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "Couldn't parse token",
		})
		return
	}

	// Check expire time of the given token
	if claims.ExpiresAt.Unix() < time.Now().Local().Unix() {
		logger.Errorf("refresh_token is expired")
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":   "token expired",
			"message": "Token is expired",
		})
		return
	}

	// Again validate token
	if !tkn.Valid {
		logger.Errorf("invalid token")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid token",
			"message": "The token is invalid",
		})
		return
	}

	// Create refresh token
	tokens, err := middlewares.RefreshToken(claims)
	if err != nil {
		logger.Errorf("couldn't create refresh token")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "Couldn't create refresh token",
		})
		return
	}

	tokens.RefreshToken = token.RefreshToken
	ctx.JSON(http.StatusOK, tokens)
}
