package models

import (
	"time"

	"github.com/google/uuid"
)

type RangeBody struct {
	APIKey string `json:"api_key" binding:"required"`
	Detail string `json:"detail"`
	Note   string `json:"note"`
}

type RangeBodyResponse struct {
	UUID        uuid.UUID  `json:"uuid"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Client      string     `json:"client"`
	Detail      string     `json:"detail"`
	Note        string     `json:"note"`
	TotalAmount float64    `json:"total_amount"`
	Currencies  Currencies `json:"currencies"`
}
