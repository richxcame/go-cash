package controllers

import (
	"encoding/json"
	"gocash/config"
	"gocash/models"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CheckBookingNumber(ctx *gin.Context) {
	var bookingsResponse models.BookingsResponse

	ticketNumber := ctx.Param("booking-number")
	url := config.GlobalConfig.BOOKINGS_API_URL + ticketNumber

	response, err := http.Get(url)
	if err != nil {
		ctx.JSON(400, err.Error())
		return
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		ctx.JSON(400, err.Error())
		return
	}

	if err := json.Unmarshal(responseBody, &bookingsResponse); err != nil {
		ctx.JSON(400, err.Error())
		return
	}

	if !bookingsResponse.Success {
		ctx.JSON(200, gin.H{"success": true, "message": "not found"})
		return
	}

	ctx.JSON(200, bookingsResponse)
}
