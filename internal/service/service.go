package service

import (
	"crypto/sha256"
	"encoding/base64"
	"sync"

	"github.com/rutkin/gofermart/internal/config"
	"github.com/rutkin/gofermart/internal/logger"
	"github.com/rutkin/gofermart/internal/models"
	"github.com/rutkin/gofermart/internal/repository"
	"go.uber.org/zap"
)

func NewService(config *config.Config) (*Service, error) {
	db, err := repository.NewDatabase(config.DatabaseURI)
	if err != nil {
		return nil, err
	}
	ls := NewLoyaltySystem(config.AccrualSystemAddress)
	return &Service{db, ls, sync.WaitGroup{}}, nil
}

type Service struct {
	db *repository.Database
	ls *LoyaltySystem
	wg sync.WaitGroup
}

func calculateHash(value string) string {
	h := sha256.New()
	h.Write([]byte(value))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func (s *Service) processOrders(userID string, ordersNumbers []string) {
	defer s.wg.Done()
	for _, orderNumber := range ordersNumbers {
		orderInfo, err := s.ls.GetOrdersInfo(orderNumber)
		if err != nil {
			logger.Log.Error("failed to get order info from loyalty system", zap.String("error", err.Error()))
			return
		}
		s.db.UpdateOrder(userID, orderInfo.Number, orderInfo.Status, orderInfo.Accrual)
	}
}

func (s *Service) Close() {
	s.wg.Wait()
}

func (s *Service) RegisterUser(username string, password string) (string, error) {
	return s.db.CreateUser(username, calculateHash(password))
}

func (s *Service) Login(username string, password string) (string, error) {
	return s.db.GetUserID(username, calculateHash(password))
}

func (s *Service) CreateOrder(userID string, orderNumber string) error {
	logger.Log.Info("create order", zap.String("number", orderNumber))
	return s.db.CreateOrder(userID, orderNumber)
}

func (s *Service) GetOrders(userID string) (models.OrdersResponse, error) {
	orders, err := s.db.GetOrders(userID)
	if err != nil {
		return models.OrdersResponse{}, err
	}
	var orderNumbersToProcess []string
	for key, order := range orders {
		if order.Status == "NEW" {
			orders[key].Status = "PROCESSING"
			logger.Log.Info("get order", zap.String("number", order.Number))
			orderNumbersToProcess = append(orderNumbersToProcess, order.Number)
		}
	}
	s.wg.Add(1)
	go s.processOrders(userID, orderNumbersToProcess)
	return orders, nil
}

func (s *Service) GetBalance(userID string) (models.BalanceRecord, error) {
	return models.BalanceRecord{}, nil
}
