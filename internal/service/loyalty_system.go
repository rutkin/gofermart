package service

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	myerrors "github.com/rutkin/gofermart/internal/errors"
	"github.com/rutkin/gofermart/internal/logger"
	"github.com/rutkin/gofermart/internal/models"
	"go.uber.org/zap"
)

var myClient = &http.Client{Timeout: 10 * time.Second}

func NewLoyaltySystem(address string) *LoyaltySystem {
	return &LoyaltySystem{address, make(chan bool)}
}

type LoyaltySystem struct {
	address     string
	stopProcess chan bool
}

func (ls *LoyaltySystem) Stop() {
	ls.stopProcess <- true
}

func (ls *LoyaltySystem) GetOrdersInfo(orderNumber string) (models.LoyaltyOrderRecord, error) {
	address := ls.address + "/api/orders/" + orderNumber
	logger.Log.Info("get order info", zap.String("address", address))
	var resp *http.Response
loop:
	for {
		select {
		case <-ls.stopProcess:
			logger.Log.Info("stop process order", zap.String("number", orderNumber))
			return models.LoyaltyOrderRecord{}, myerrors.ErrTimeout
		default:
			var err error
			resp, err = myClient.Get(address)
			if err != nil {
				logger.Log.Error("failed to get order info from loyalty system", zap.String("error", err.Error()))
				return models.LoyaltyOrderRecord{}, err
			}
			switch resp.StatusCode {
			case http.StatusTooManyRequests:
				timeout, err := strconv.Atoi(resp.Header.Get("Retry-After"))
				if err != nil {
					logger.Log.Error("failed to convert retry from header", zap.String("error", err.Error()))
					return models.LoyaltyOrderRecord{}, myerrors.ErrInternal
				}
				time.Sleep(time.Duration(timeout) * time.Second)
			case http.StatusNoContent:
				logger.Log.Info("no content")
				time.Sleep(1 * time.Second)
			case http.StatusInternalServerError:
				logger.Log.Error("internal server error")
				return models.LoyaltyOrderRecord{}, myerrors.ErrInternal
			default:
				break loop
			}
		}
	}

	var loyaltyOrder models.LoyaltyOrderRecord
	if err := json.NewDecoder(resp.Body).Decode(&loyaltyOrder); err != nil {
		logger.Log.Error("failed to decode loyalty order", zap.String("error", err.Error()))
		body, _ := io.ReadAll(resp.Body)
		logger.Log.Error("response", zap.String("body", string(body)))
	}
	defer resp.Body.Close()
	return loyaltyOrder, nil
}
