package middleware

import (
	"context"
	"net/http"

	myerrors "github.com/rutkin/gofermart/internal/errors"
	"github.com/rutkin/gofermart/internal/helpers"
)

func GetUserIDFromCookie(r *http.Request) (string, error) {
	userIDCookie, err := r.Cookie(helpers.UserIDKey)
	if err != nil {
		return "", myerrors.ErrNotFound
	}

	userID, err := helpers.Decode(userIDCookie.Value)
	return userID, err
}

func WithAuth(h http.Handler) http.Handler {
	authFn := func(w http.ResponseWriter, r *http.Request) {
		userID, err := GetUserIDFromCookie(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), helpers.UserIDKey, userID)
		h.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(authFn)
}
