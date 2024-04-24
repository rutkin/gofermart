package service

import (
	"encoding/json"
	"net/http"

	"github.com/rutkin/gofermart/internal/logger"
	"github.com/rutkin/gofermart/internal/models"
	"go.uber.org/zap"
)

func NewLoyaltySystem(address string) *LoyaltySystem {
	return &LoyaltySystem{address}
}

type LoyaltySystem struct {
	address string
}

func (ls *LoyaltySystem) GetOrdersInfo(orderNumber string) (models.LoyaltyOrderRecord, error) {
	resp, err := http.Get(ls.address + "api/users/" + orderNumber)
	var loyaltyOrder models.LoyaltyOrderRecord
	if err := json.NewDecoder(resp.Body).Decode(&loyaltyOrder); err != nil {
		logger.Log.Error("failed to decode loyalty order", zap.String("error", err.Error()))
	}
	return loyaltyOrder, err
}
