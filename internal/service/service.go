package service

import (
	"crypto/sha256"
	"encoding/base64"

	"github.com/rutkin/gofermart/internal/config"
	"github.com/rutkin/gofermart/internal/repository"
)

func NewService(config *config.Config) (*Service, error) {
	db, err := repository.NewDatabase(config.DatabaseURI)
	if err != nil {
		return nil, err
	}
	return &Service{db}, nil
}

type Service struct {
	db *repository.Database
}

func calculateHash(value string) string {
	h := sha256.New()
	h.Write([]byte(value))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func (s *Service) RegisterUser(username string, password string) (string, error) {
	return s.db.CreateUser(username, calculateHash(password))
}

func (s *Service) Login(username string, password string) (string, error) {
	return s.db.GetUserID(username, calculateHash(password))
}

func (s *Service) CreateOrder(username string, orderNumber string) error {
	return nil
}
