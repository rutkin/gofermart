package service

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rutkin/gofermart/internal/logger"
	"github.com/rutkin/gofermart/internal/models"
	"go.uber.org/zap"
)

var myClient = &http.Client{Timeout: 10 * time.Second}

func NewLoyaltySystem(address string) *LoyaltySystem {
	return &LoyaltySystem{address}
}

type LoyaltySystem struct {
	address string
}

func (ls *LoyaltySystem) GetOrdersInfo(orderNumber string) (models.LoyaltyOrderRecord, error) {
	resp, err := myClient.Get(ls.address + "api/users/" + orderNumber)
	var loyaltyOrder models.LoyaltyOrderRecord
	if err := json.NewDecoder(resp.Body).Decode(&loyaltyOrder); err != nil {
		logger.Log.Error("failed to decode loyalty order", zap.String("error", err.Error()))
	}
	defer resp.Body.Close()
	return loyaltyOrder, err
}
