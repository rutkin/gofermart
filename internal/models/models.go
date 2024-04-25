package models

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type OrderRecord struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float32 `json:"accrual,omitempty"`
	UploadetAt string  `json:"uploadet_at"`
}

type OrdersResponse []OrderRecord

type LoyaltyOrderRecord struct {
	Number  string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual,omitempty"`
}

type BalanceRecord struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}
