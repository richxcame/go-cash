package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gocash/config"
	"gocash/models"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CheckBookingNumber(ctx *gin.Context) {
	var response models.CheckBookingResponse
	bookingResponse := checkBookingFromApi(ctx, &response)
	checkIsMoneySent(&response, bookingResponse)
	ctx.JSON(200, response)
}

func checkBookingFromApi(ctx *gin.Context, checkBooking *models.CheckBookingResponse) models.BookingsResponse {
	var bookingsResponse models.BookingsResponse

	ticketNumber := ctx.Param("booking-number")
	url := config.GlobalConfig.BOOKINGS_API_URL + ticketNumber

	response, err := http.Get(url)
	if err != nil {
		ctx.JSON(400, err.Error())
		return models.BookingsResponse{}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		ctx.JSON(400, err.Error())
		return models.BookingsResponse{}
	}

	if err := json.Unmarshal(responseBody, &bookingsResponse); err != nil {
		ctx.JSON(400, err.Error())
		return models.BookingsResponse{}
	}

	if !bookingsResponse.Success {
		checkBooking.Booking.Message = "not found"
		checkBooking.Booking.Success = false
		return models.BookingsResponse{}
	}

	checkBooking.Booking.Message = "found"
	checkBooking.Booking.Success = true
	return bookingsResponse
}

func checkIsMoneySent(checkBooking *models.CheckBookingResponse, bookingResponse models.BookingsResponse) {
	token, err := LoginToGotoleg()
	if err != nil {
		checkBooking.Transaction.Message = err.Error()
		checkBooking.Transaction.Success = false
		return
	}
	err = checkTransaction(token, bookingResponse)
	if err != nil {
		checkBooking.Transaction.Message = err.Error()
		checkBooking.Transaction.Success = false
		return
	}
	checkBooking.Transaction.Message = "success"
	checkBooking.Transaction.Success = true
}

func checkTransaction(token models.GotolegToken, bookingResponse models.BookingsResponse) error {
	var transactionsResponse models.GotolegResponseTransactions
	client := &http.Client{}

	// Create a new GET request
	req, err := http.NewRequest("GET", config.GlobalConfig.GOTOLEG_URL+"transactions?note="+bookingResponse.Data.Booking.BookingNumber, nil)
	if err != nil {

		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Add("User-Agent", "MyGoApp/1.0")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending GET request:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %s", resp.Status)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(responseBody, &transactionsResponse); err != nil {
		return err
	}

	if len(transactionsResponse.Transactions) == 0 {
		return errors.New("not found")
	}

	if transactionsResponse.Transactions[0].Status != "SUCCESS" || transactionsResponse.Transactions[0].ResultStatus != "SUCCESS" {
		return errors.New("not refunded")
	}

	return nil
}

func LoginToGotoleg() (models.GotolegToken, error) {
	var token models.GotolegToken
	payload := []byte(fmt.Sprintf(`{"username":"%s", "password": "%s"}`, config.GlobalConfig.GOTOLEG_LOGIN, config.GlobalConfig.GOTOLEG_PASS))

	req, err := http.NewRequest("POST", config.GlobalConfig.GOTOLEG_URL+"auth/login", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return models.GotolegToken{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return models.GotolegToken{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return models.GotolegToken{}, fmt.Errorf("request failed with status: %s", resp.Status)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.GotolegToken{}, err
	}

	if err := json.Unmarshal(responseBody, &token); err != nil {
		return models.GotolegToken{}, err
	}

	return token, nil
}
