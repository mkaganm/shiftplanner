package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"shiftplanner/backend/internal/database"
	"shiftplanner/backend/internal/models"
	"time"
)

// CreateUser creates a new user
func CreateUser(username, password string) (*models.User, error) {
	// Hash password
	passwordHash := hashPassword(password)

	result, err := database.DB.Exec(
		"INSERT INTO users (username, password_hash) VALUES (?, ?)",
		username, passwordHash,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:        int(id),
		Username:  username,
		CreatedAt: time.Now(),
	}, nil
}

// GetUserByUsername gets a user by username
func GetUserByUsername(username string) (*models.User, string, error) {
	var user models.User
	var passwordHash string
	var createdAtStr string

	err := database.DB.QueryRow(
		"SELECT id, username, password_hash, created_at FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &passwordHash, &createdAtStr)

	if err != nil {
		return nil, "", err
	}

	user.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	return &user, passwordHash, nil
}

// ValidatePassword validates a password
func ValidatePassword(password, passwordHash string) bool {
	hashed := hashPassword(password)
	return hashed == passwordHash
}

// hashPassword hashes a password
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}
