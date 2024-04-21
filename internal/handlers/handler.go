package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rutkin/gofermart/internal/config"
	my_errors "github.com/rutkin/gofermart/internal/errors"
	"github.com/rutkin/gofermart/internal/logger"
	"github.com/rutkin/gofermart/internal/models"
	"github.com/rutkin/gofermart/internal/service"
	"go.uber.org/zap"
)

func getRegisterRequest(r *http.Request) (models.RegisterRequest, error) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Error("failed to decode body", zap.String("error", err.Error()))
		return req, err
	}
	return req, nil
}

func NewHandler(config *config.Config) (*Handler, error) {
	service, err := service.NewService(config)
	if err != nil {
		return nil, err
	}
	return &Handler{service}, nil
}

type Handler struct {
	service *service.Service
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	req, err := getRegisterRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := h.service.RegisterUser(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, my_errors.ErrExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	userIDcookie := &http.Cookie{Name: "userID", Value: userID}
	http.SetCookie(w, userIDcookie)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	req, err := getRegisterRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := h.service.Login(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, my_errors.ErrNotFound) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}
	userIDcookie := &http.Cookie{Name: "userID", Value: userID}
	http.SetCookie(w, userIDcookie)
	w.WriteHeader(http.StatusOK)
}
