package models

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type OrderRecord struct {
	Number     string `json:"number"`
	Status     string `json:"status"`
	Accrual    int    `json:"accrual,omitempty"`
	UploadetAt string `json:"uploadet_at"`
}

type OrdersResponse []OrderRecord

type LoyaltyOrderRecord struct {
	Number  string `json:"number"`
	Status  string `json:"status"`
	Accrual int32  `json:"accrual,omitempty"`
}
