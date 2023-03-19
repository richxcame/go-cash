package main

import (
	"context"
	"gocash/pkg/db"
	"gocash/pkg/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
)

type CashBody struct {
	APIKey  string  `json:"api_key" binding:"required"`
	Amount  float64 `json:"amount" binding:"required"`
	Contact string  `json:"contact" binding:"required"`
	Detail  string  `json:"detail"`
	Note    string  `json:"note"`
}

type RangeBody struct {
	APIKey string `json:"api_key" binding:"required"`
	Detail string `json:"detail"`
	Note   string `json:"note"`
}

func main() {
	// Database instance
	db := db.CreateDB()
	defer db.Close()

	r := gin.Default()

	r.POST("/cashes", func(ctx *gin.Context) {
		// Get request body
		var body CashBody
		if err := ctx.BindJSON(&body); err != nil {
			logger.Errorf("request body wrong %v", err)
			ctx.JSON(400, gin.H{
				"error":   err.Error(),
				"message": "Request body invalid",
			})
			return
		}

		// Find the client with the given key
		var client string
		err := db.QueryRow(context.Background(), "SELECT name FROM clients WHERE api_key = $1", body.APIKey).Scan(&client)
		if err != nil {
			logger.Errorf("api key search error %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err,
				"message": "Client hasn't been found",
			})
			return
		}

		// Insert request to database
		_uuid := uuid.New().String()
		sqlStatement := `
		INSERT INTO cashes (uuid, created_at, updated_at, client, contact, amount, detail, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err = db.Exec(context.Background(), sqlStatement, _uuid, time.Now(), time.Now(), client, body.Contact, body.Amount, body.Detail, body.Note)
		if err != nil {
			logger.Errorf("database save error %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err,
				"message": "Couldn't save into cashes",
			})
			return
		}

		// Send success result
		ctx.JSON(201, gin.H{
			"message": "Successfully saved into database",
			"uuid":    _uuid,
		})
	})

	r.POST("/ranges", func(ctx *gin.Context) {
		// Get request body
		var body RangeBody
		if err := ctx.BindJSON(&body); err != nil {
			logger.Errorf("request body wrong %v", err)
			ctx.JSON(400, gin.H{
				"error":   err.Error(),
				"message": "Request body invalid",
			})
			return
		}

		// Find the client with the given key
		var client string
		err := db.QueryRow(context.Background(), "SELECT name FROM clients WHERE api_key = $1", body.APIKey).Scan(&client)
		if err != nil {
			logger.Errorf("api key search error %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err,
				"message": "Client hasn't been found",
			})
			return
		}

		// Insert request to database
		_uuid := uuid.New().String()
		sqlStatement := `
		INSERT INTO ranges (uuid, created_at, updated_at, client, detail, note)
		VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = db.Exec(context.Background(), sqlStatement, _uuid, time.Now(), time.Now(), client, body.Detail, body.Note)
		if err != nil {
			logger.Errorf("database save error %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err,
				"message": "Couldn't save into cashes",
			})
			return
		}

		// Send success result
		ctx.JSON(201, gin.H{
			"message": "Successfully saved into database",
			"uuid":    _uuid,
		})
	})

	r.Run()
}
