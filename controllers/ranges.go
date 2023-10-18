package controllers

import (
	"context"
	"gocash/models"
	"gocash/pkg/db"
	"gocash/pkg/logger"
	"gocash/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func RangesController(ctx *gin.Context) {
	// Get request body
	var body models.RangeBody
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
	err := db.DB.QueryRow(context.Background(), "SELECT name FROM clients WHERE api_key = $1", body.APIKey).Scan(&client)
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
	_, err = db.DB.Exec(context.Background(), sqlStatement, _uuid, time.Now(), time.Now(), client, body.Detail, body.Note)
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
}

func RangesWithAuth(ctx *gin.Context) {
	offset, limit := utils.Paginate(ctx)

	// Find ranges
	var rangeBodies []models.RangeBodyResponse
	sqlStatement := `SELECT r.uuid, r.created_at, r.updated_at, r.client, r.detail, r.note FROM ranges r ORDER BY r.created_at DESC OFFSET $1 LIMIT $2;`
	rows, err := db.DB.Query(context.Background(), sqlStatement, offset, limit)
	if err != nil {
		logger.Errorf("ranges search error %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "Something went wrong",
		})
		return
	}
	for rows.Next() {
		var rangeBody models.RangeBodyResponse
		err := rows.Scan(&rangeBody.UUID, &rangeBody.CreatedAt, &rangeBody.UpdatedAt, &rangeBody.Client, &rangeBody.Detail, &rangeBody.Note)
		if err != nil {
			logger.Errorf("Scan error %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "Scan error",
			})
			return
		}
		rangeBodies = append(rangeBodies, rangeBody)
	}

	resultRanges := make([]models.RangeBodyResponse, 0)
	for _, v := range rangeBodies {
		var rangeBody models.RangeBodyResponse
		err := db.DB.QueryRow(context.Background(), "SELECT r.created_at, r.client FROM ranges r  WHERE created_at < $1 AND client=$2 ORDER BY created_at DESC limit 1", v.CreatedAt, v.Client).Scan(&rangeBody.CreatedAt, &rangeBody.Client)
		if err != nil && err != pgx.ErrNoRows {
			logger.Errorf("last range not found: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "Couldn't find searching range",
			})
			return
		} else if err == pgx.ErrNoRows {
			rangeBody.CreatedAt = time.Date(2001, 12, 28, 0, 0, 0, 0, time.Now().Location())
		}

		// Search sum of cashes between the two ranges
		var totalAmount *float64
		err = db.DB.QueryRow(context.Background(), "SELECT SUM(amount) FROM cashes where created_at >= $1 AND created_at <= $2 AND client=$3", rangeBody.CreatedAt, v.CreatedAt, v.Client).Scan(&totalAmount)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "Couldn't find total amount of the range",
			})
		}
		if totalAmount == nil {
			defaultTotalAmount := float64(0)
			totalAmount = &defaultTotalAmount
		}

		// Search sum and amount of one currency cashes between the two ranges
		var currencies models.Currencies
		for _, vCurrency := range []uint{1, 5, 10, 20, 50, 100} {
			var currency models.Currency
			rowOne, err := db.DB.Query(context.Background(), "SELECT SUM(amount),COUNT(amount) FROM cashes where created_at >= $1 AND created_at <= $2 AND client=$3 AND amount = $4 GROUP BY amount", rangeBody.CreatedAt, v.CreatedAt, v.Client, vCurrency)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error":   err.Error(),
					"message": "Couldn't find total amount of the range",
				})
			}

			for rowOne.Next() {
				err := rowOne.Scan(&currency.TotalAmount, &currency.Amount)
				if err != nil {
					logger.Errorf("Scan error %v", err)
					ctx.JSON(http.StatusInternalServerError, gin.H{
						"error":   err.Error(),
						"message": "Scan error",
					})
					return
				}
			}

			switch vCurrency {
			case 1:
				currencies.One = currency
			case 5:
				currencies.Five = currency
			case 10:
				currencies.Ten = currency
			case 20:
				currencies.Twenty = currency
			case 50:
				currencies.Fifty = currency
			case 100:
				currencies.OneHundred = currency
			}
		}

		resultRanges = append(resultRanges, models.RangeBodyResponse{
			UUID:        v.UUID,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
			Client:      v.Client,
			Note:        v.Note,
			Detail:      v.Detail,
			TotalAmount: *totalAmount,
			Currencies:  currencies,
		})
	}

	// Find total count of ranges
	totalRanges := 0
	err = db.DB.QueryRow(context.Background(), "SELECT COUNT(*) FROM ranges").Scan(&totalRanges)
	if err != nil {
		logger.Errorf("couldn't find total count of ranges: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "Couldn't find total number of ranges",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"ranges": resultRanges,
		"total":  totalRanges,
	})
}
