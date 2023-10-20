package models

import "time"

type BookingsResponse struct {
	Success bool `json:"success"`
	Data    Data `json:"data"`
}
type Passenger struct {
	Name           string `json:"name"`
	Surname        string `json:"surname"`
	Dob            string `json:"dob"`
	Gender         string `json:"gender"`
	IdentityType   string `json:"identity_type"`
	IdentityNumber string `json:"identity_number"`
}
type PriceFormation struct {
	ID     int     `json:"id"`
	Title  string  `json:"title"`
	Amount float64 `json:"amount"`
}
type Pnrs struct {
	ID               int              `json:"id"`
	TicketNumber     string           `json:"ticket_number"`
	TrainRunNumber   string           `json:"train_run_number"`
	Source           string           `json:"source"`
	Destination      string           `json:"destination"`
	DepartureTime    time.Time        `json:"departure_time"`
	ArrivalTime      time.Time        `json:"arrival_time"`
	ServiceTypeTitle string           `json:"service_type_title"`
	WagonTypeTitle   string           `json:"wagon_type_title"`
	WagonNumber      int              `json:"wagon_number"`
	SeatLabel        string           `json:"seat_label"`
	ReturnTicket     int              `json:"return_ticket"`
	Status           int              `json:"status"`
	PriceFormation   []PriceFormation `json:"price_formation"`
}
type Tickets struct {
	QrCode    string    `json:"qr_code"`
	PdfURL    string    `json:"pdf_url"`
	Passenger Passenger `json:"passenger"`
	Pnrs      []Pnrs    `json:"pnrs"`
}
type Payments struct {
	PaymentType string    `json:"payment_type"`
	Amount      float64   `json:"amount"`
	Receipt     string    `json:"receipt"`
	PaymentDate time.Time `json:"payment_date"`
	Details     string    `json:"details"`
}
type Booking struct {
	BookingNumber string     `json:"booking_number"`
	MainContact   string     `json:"main_contact"`
	Phone         string     `json:"phone"`
	Email         any        `json:"email"`
	TotalPrice    float64    `json:"total_price"`
	ExpireTime    time.Time  `json:"expire_time"`
	Tickets       []Tickets  `json:"tickets"`
	Payments      []Payments `json:"payments"`
}
type Data struct {
	Booking Booking `json:"booking"`
}

type GotolegToken struct {
	AccessToken string `json:"access_token"`
}

type CheckBookingResponse struct {
	Booking struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	} `json:"booking"`
	Transaction struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	} `json:"transaction"`
}

type GotolegResponseTransactions struct {
	Limit        int            `json:"limit"`
	Offset       int            `json:"offset"`
	Statuses     []string       `json:"statuses"`
	TotalCount   int            `json:"total_count"`
	Transactions []Transactions `json:"transactions"`
}
type Transactions struct {
	UUID              string    `json:"uuid"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	RequestLocalID    string    `json:"request_local_id"`
	RequestService    string    `json:"request_service"`
	RequestPhone      string    `json:"request_phone"`
	RequestAmount     string    `json:"request_amount"`
	Status            string    `json:"status"`
	ErrorCode         int       `json:"error_code"`
	ErrorMsg          string    `json:"error_msg"`
	ResultStatus      string    `json:"result_status"`
	ResultRefNum      string    `json:"result_ref_num"`
	ResultService     string    `json:"result_service"`
	ResultDestination string    `json:"result_destination"`
	ResultAmount      int       `json:"result_amount"`
	ResultState       string    `json:"result_state"`
	ResultReason      string    `json:"result_reason"`
	IsChecked         bool      `json:"is_checked"`
	Client            string    `json:"client"`
	Note              string    `json:"note"`
}
