package middleware

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"

	myerrors "github.com/rutkin/gofermart/internal/errors"
	"github.com/rutkin/gofermart/internal/helpers"
	"github.com/rutkin/gofermart/internal/logger"
	"go.uber.org/zap"
)

const password = "password"

func GetUserIDFromCookie(r *http.Request) (string, error) {
	userIDCookie, err := r.Cookie(helpers.UserIDKey)
	if err != nil {
		return "", myerrors.ErrNotFound
	}

	key := sha256.Sum256([]byte(password))
	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		logger.Log.Error("failed to create new cipher", zap.String("error", err.Error()))
		return "", err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		logger.Log.Error("failed to create new gcm", zap.String("error", err.Error()))
		return "", err
	}

	nonce := key[(len(key) - aesgcm.NonceSize()):]

	data, err := hex.DecodeString(userIDCookie.Value)
	if err != nil {
		logger.Log.Error("failed to decode userID cookie", zap.String("error", err.Error()))
		return "", err
	}

	userID, err := aesgcm.Open(nil, nonce, data, nil)
	if err != nil {
		logger.Log.Error("failed to decrypt userID cookie", zap.String("error", err.Error()))
		return "", err
	}
	return string(userID), err
}

func WithAuth(h http.Handler) http.Handler {
	authFn := func(w http.ResponseWriter, r *http.Request) {
		userID, err := GetUserIDFromCookie(r)
		if err != nil {
			if errors.Is(err, myerrors.ErrNotFound) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), helpers.UserIDKey, userID)
		h.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(authFn)
}
