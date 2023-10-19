package models

import (
	"time"

	"github.com/google/uuid"
)

type CashBody struct {
	APIKey  string  `json:"api_key" binding:"required"`
	Amount  float64 `json:"amount" binding:"required"`
	Contact string  `json:"contact" binding:"required"`
	Detail  string  `json:"detail"`
	Note    string  `json:"note"`
}

type CashBodyResponseBooking struct {
	UUID      uuid.UUID `json:"uuid"`
	Amount    float64   `json:"amount" binding:"required"`
	Detail    string    `json:"detail"`
	Note      string    `json:"note"`
	Client    string    `json:"client"`
	Contact   string    `json:"contact"`
	CreatedAt time.Time `json:"created_at"`
}

type CashBodyResponse struct {
	Amount    float64   `json:"amount" binding:"required"`
	Detail    string    `json:"detail"`
	Contact   string    `json:"contact"`
	CreatedAt time.Time `json:"created_at"`
}
