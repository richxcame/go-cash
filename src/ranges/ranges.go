package ranges

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
)

func CreateRangesController(ctx *gin.Context) {
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
	err := db.DB.QueryRow(context.Background(), slClSql, body.APIKey).Scan(&client)
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
	_, err = db.DB.Exec(context.Background(), inRngsSql, _uuid, time.Now(), time.Now(), client, body.Detail, body.Note)
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

func GetRanges(ctx *gin.Context) {
	offset, limit := utils.Paginate(ctx)
	resultRanges, totalRanges, errMsh, msg := GetRangesService(offset, limit)
	if errMsh != "" {
		logger.Errorf(msg+" %v", errMsh)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   errMsh,
			"message": msg,
		})
	}
	ctx.JSON(200, gin.H{
		"ranges": resultRanges,
		"total":  totalRanges,
	})
}
