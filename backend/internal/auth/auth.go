package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"shiftplanner/backend/internal/database"
	"shiftplanner/backend/internal/models"
	"time"
)

const (
	TokenLength   = 32
	SessionExpiry = 7 * 24 * time.Hour // 7 days
)

// GenerateToken generates a random token
func GenerateToken() (string, error) {
	bytes := make([]byte, TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateSession creates a new session
func CreateSession(userID int) (*models.Session, error) {
	token, err := GenerateToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(SessionExpiry)

	result, err := database.DB.Exec(
		"INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, ?)",
		userID, token, expiresAt,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.Session{
		ID:        int(id),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}, nil
}

// ValidateToken validates the token and returns the user ID
func ValidateToken(token string) (int, error) {
	var userID int
	var expiresAt time.Time

	err := database.DB.QueryRow(
		"SELECT user_id, expires_at FROM sessions WHERE token = ?",
		token,
	).Scan(&userID, &expiresAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrInvalidToken
		}
		return 0, err
	}

	if time.Now().After(expiresAt) {
		// Delete expired session
		database.DB.Exec("DELETE FROM sessions WHERE token = ?", token)
		return 0, ErrExpiredToken
	}

	return userID, nil
}

// DeleteSession deletes a session
func DeleteSession(token string) error {
	_, err := database.DB.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}

// CleanExpiredSessions cleans expired sessions
func CleanExpiredSessions() error {
	_, err := database.DB.Exec("DELETE FROM sessions WHERE expires_at < ?", time.Now())
	return err
}

// Error definitions
var (
	ErrInvalidToken = &AuthError{Message: "Invalid token"}
	ErrExpiredToken = &AuthError{Message: "Session expired"}
)

// AuthError authentication error
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}
