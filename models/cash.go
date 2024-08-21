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
type Reason struct {
	Reason string  `json:"reason"`
	Sum    float64 `json:"sum"`
}

type Toleg struct {
	TotalSum float64  `json:"total_sum"`
	Reasons  []Reason `json:"reasons"`
}

type ResultDatesCashes struct {
	TotalCash  float64 `json:"total_cash"`
	TotalToleg Toleg   `json:"total_toleg"`
}
