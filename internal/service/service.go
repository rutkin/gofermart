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

func (s *Service) processOrder(userID string, orderNumber string) {
	defer s.wg.Done()
	s.db.UpdateOrder(userID, orderNumber, "PROCESSING", 0)
	orderInfo, err := s.ls.GetOrdersInfo(orderNumber)
	if err != nil {
		logger.Log.Error("failed to get order info from loyalty system", zap.String("error", err.Error()))
		return
	}
	logger.Log.Info("update order", zap.String("userID", userID), zap.String("number", orderInfo.Number), zap.String("syayus", orderInfo.Status), zap.Float32("accrual", orderInfo.Accrual))
	s.db.UpdateOrder(userID, orderInfo.Number, orderInfo.Status, orderInfo.Accrual)
}

func (s *Service) Close() {
	s.ls.Stop()
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
	err := s.db.CreateOrder(userID, orderNumber)
	if err == nil {
		s.wg.Add(1)
		go s.processOrder(userID, orderNumber)
	}
	return err
}

func (s *Service) GetOrders(userID string) (models.OrdersResponse, error) {
	orders, err := s.db.GetOrders(userID)
	if err != nil {
		return models.OrdersResponse{}, err
	}

	return orders, nil
}

func (s *Service) GetBalance(userID string) (models.BalanceRecord, error) {
	return s.db.GetBalance(userID)
}

func (s *Service) Withdraw(userID string, rec models.WithdrawRecord) error {
	/*_, err := s.db.GetOrder(rec.Number)
	if err != nil {
		logger.Log.Info("failed to get order", zap.String("error", err.Error()))
		return err
	}*/

	err := s.db.Withdraw(userID, rec.Sum)
	if err != nil {
		logger.Log.Error("failed to withdraw", zap.String("error", err.Error()))
	}
	return nil
}
