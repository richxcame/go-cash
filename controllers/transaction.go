package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gocash/config"
	"gocash/models"
	"gocash/pkg/db"
	"gocash/pkg/logger"
	"io"
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

	payload, err := json.Marshal(transaction)
	if err != nil {
		logger.Errorf("couldn't marshal json  %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err,
			"message": "couldn't marshal json",
		})
		return
	}

	if err := sendRequest(payload); err != nil {
		logger.Errorf("error in sending refund  %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err,
			"message": "error in sending refund",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "Successfully refunded",
	})
}

func sendRequest(payload []byte) error {
	req, err := http.NewRequest("POST", config.GlobalConfig.GOTOLEG_URL+"trnxs", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %s", resp.Status)
	}

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
}