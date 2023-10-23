package controllers

import (
	"gocash/config"
	"gocash/models"
	"gocash/pkg/db"
	"gocash/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Transaction(ctx *gin.Context) {
	var sendMoneyRequest models.SendMoneyRequest

	if err := ctx.BindJSON(&sendMoneyRequest); err != nil {
		logger.Errorf("data parse error %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err,
			"message": "invalid data",
		})
	}
	var transaction models.SendMoney
	query := `SELECT contact from cashes WHERE detail = $1`
	if err := db.DB.QueryRow(ctx, query, sendMoneyRequest.BookingNumber).
		Scan(&transaction.Phone); err != nil {
		logger.Errorf("booking number search error %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err,
			"message": "Booking number hasn't been found",
		})
		return
	}

	transaction.Note = sendMoneyRequest.BookingNumber
	transaction.APIKey = config.GlobalConfig.GOTOLEG_API_KEY
	transaction.Service = ""
	transaction.Amount = strconv.Itoa(sendMoneyRequest.Amount)
}
