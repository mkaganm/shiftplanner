package api

import (
	"context"
	"net/http"
	"shiftplanner/backend/internal/auth"
)

type contextKey string

const userIDContextKey contextKey = "userID"

// AuthMiddleware authentication middleware
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, err := auth.ValidateToken(token)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add UserID to request context
		ctx := context.WithValue(r.Context(), userIDContextKey, userID)
		next(w, r.WithContext(ctx))
	}
}

// GetUserID gets user ID from request context
func GetUserID(r *http.Request) int {
	if userID, ok := r.Context().Value(userIDContextKey).(int); ok {
		return userID
	}
	return 0
}
