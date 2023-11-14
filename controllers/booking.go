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
	bookingResponse := CheckBookingFromApi(ctx)

	if !bookingResponse.Success {
		response.Booking.Message = "not found"
		response.Booking.Success = false
	} else {
		response.Booking.Message = "found"
		response.Booking.Success = true
	}

	if err := checkIsMoneySent(bookingResponse); err != nil {
		response.Transaction.Message = err.Error()
		response.Transaction.Success = false
		ctx.JSON(200, response)
		return
	}

	response.Transaction.Message = "success"
	response.Transaction.Success = true

	ctx.JSON(200, response)
}

func CheckBookingFromApi(ctx *gin.Context) models.BookingsResponse {
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

	return bookingsResponse
}

func checkIsMoneySent(bookingResponse models.BookingsResponse) error {
	token, err := LoginToGotoleg()
	if err != nil {
		return err
	}
	err = checkTransaction(token, bookingResponse)
	if err != nil {
		return err
	}

	return nil
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
