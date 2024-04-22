package helpers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"

	"github.com/rutkin/gofermart/internal/logger"
	"go.uber.org/zap"
)

var UserIDKey = "userID"

type contextKey string

var UserIDContextKey contextKey = contextKey(UserIDKey)

const password = "password"

func Encode(value string) (string, error) {
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

	encryptedValue := aesgcm.Seal(nil, nonce, []byte(value), nil)
	return string(encryptedValue), nil
}

func Decode(value string) (string, error) {
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

	data, err := hex.DecodeString(value)
	if err != nil {
		logger.Log.Error("failed to decode", zap.String("error", err.Error()))
		return "", err
	}

	decodedValue, err := aesgcm.Open(nil, nonce, data, nil)
	if err != nil {
		logger.Log.Error("failed to decrypt", zap.String("error", err.Error()))
		return "", err
	}
	return string(decodedValue), err
}
