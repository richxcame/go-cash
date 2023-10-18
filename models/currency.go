package models

type Currencies struct {
	One        Currency `json:"one"`
	Five       Currency `json:"five"`
	Ten        Currency `json:"ten"`
	Twenty     Currency `json:"twenty"`
	Fifty      Currency `json:"fifty"`
	OneHundred Currency `json:"one_hundred"`
}

type Currency struct {
	TotalAmount float64 `json:"total_amount"`
	Amount      uint    `json:"amount"`
}
