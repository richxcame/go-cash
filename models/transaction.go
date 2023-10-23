package models

type SendMoney struct {
	LocalID string `json:"local_id"`
	Service string `json:"service"`
	Phone   string `json:"phone"`
	Amount  string `json:"amount"`
	Note    string `json:"note"`
	APIKey  string `json:"api_key"`
}

type SendMoneyRequest struct {
	BookingNumber string `json:"booking_number"`
	Amount        int    `json:"amount"`
}
