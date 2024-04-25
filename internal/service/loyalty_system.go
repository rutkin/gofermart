package service

import (
	"encoding/json"
	"io"
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
	address := ls.address + "/api/orders/" + orderNumber
	logger.Log.Info("get order info", zap.String("address", address))
	resp, err := myClient.Get(address)
	var loyaltyOrder models.LoyaltyOrderRecord
	if err := json.NewDecoder(resp.Body).Decode(&loyaltyOrder); err != nil {
		logger.Log.Error("failed to decode loyalty order", zap.String("error", err.Error()))
		body, _ := io.ReadAll(resp.Body)
		logger.Log.Error("response", zap.String("body", string(body)))
	}
	defer resp.Body.Close()
	return loyaltyOrder, err
}
