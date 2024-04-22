package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/rutkin/gofermart/internal/config"
	myerrors "github.com/rutkin/gofermart/internal/errors"
	"github.com/rutkin/gofermart/internal/helpers"
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

func setUserIDCookie(userID string, w http.ResponseWriter) {
	encryptedUserID, err := helpers.Encode(userID)
	if err != nil {
		logger.Log.Error("failed to encrypt userID", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	userIDcookie := &http.Cookie{Name: helpers.UserIDKey, Value: encryptedUserID}
	http.SetCookie(w, userIDcookie)
	w.WriteHeader(http.StatusOK)
}

func getUserID(context context.Context) string {
	userID := context.Value(helpers.UserIDContextKey)
	return userID.(string)
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
		if errors.Is(err, myerrors.ErrExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	setUserIDCookie(userID, w)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	req, err := getRegisterRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := h.service.Login(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, myerrors.ErrNotFound) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}
	setUserIDCookie(userID, w)
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	orderNumber, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Error("failed to read order number in create order request", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	strOrderNumber := string(orderNumber)
	err = goluhn.Validate(strOrderNumber)
	if err != nil {
		logger.Log.Error("failed to validate order number", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	userID := getUserID(r.Context())
	err = h.service.CreateOrder(userID, strOrderNumber)
	if err != nil {
		if errors.Is(err, myerrors.ErrExists) {
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, myerrors.ErrConflict) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		logger.Log.Error("failed to create order", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
