package auth

import (
	"database/sql"
	"os"
	"path/filepath"
	"shiftplanner/backend/internal/database"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupAuthTestDB(t *testing.T) int {
	testDBPath := filepath.Join(os.TempDir(), "test_auth.db")
	os.Remove(testDBPath)

	var err error
	database.DB, err = sql.Open("sqlite3", testDBPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create schema
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	createSessionsTable := `
	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL UNIQUE,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	if _, err := database.DB.Exec(createUsersTable); err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}
	if _, err := database.DB.Exec(createSessionsTable); err != nil {
		t.Fatalf("Failed to create sessions table: %v", err)
	}

	// Create test user
	result, _ := database.DB.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", "testuser", "testhash")
	userID, _ := result.LastInsertId()
	return int(userID)
}

func teardownAuthTestDB(t *testing.T) {
	if database.DB != nil {
		database.DB.Close()
		database.DB = nil
	}
	testDBPath := filepath.Join(os.TempDir(), "test_auth.db")
	os.Remove(testDBPath)
}

func TestGenerateToken(t *testing.T) {
	token1, err := GenerateToken()
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if len(token1) != TokenLength*2 { // hex encoding doubles the length
		t.Errorf("Token length mismatch: got %d, want %d", len(token1), TokenLength*2)
	}

	// Two tokens should be different
	token2, _ := GenerateToken()
	if token1 == token2 {
		t.Error("Tokens should be different")
	}
}

func TestCreateSession(t *testing.T) {
	userID := setupAuthTestDB(t)
	defer teardownAuthTestDB(t)

	session, err := CreateSession(userID)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if session.UserID != userID {
		t.Errorf("Session user ID mismatch: got %d, want %d", session.UserID, userID)
	}

	if session.Token == "" {
		t.Error("Token cannot be empty")
	}

	if session.ExpiresAt.Before(time.Now()) {
		t.Error("ExpiresAt should be in the future")
	}
}

func TestValidateToken(t *testing.T) {
	userID := setupAuthTestDB(t)
	defer teardownAuthTestDB(t)

	// Create valid session
	session, _ := CreateSession(userID)

	// Test valid token
	validatedUserID, err := ValidateToken(session.Token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if validatedUserID != userID {
		t.Errorf("User ID mismatch: got %d, want %d", validatedUserID, userID)
	}

	// Test invalid token
	_, err = ValidateToken("invalid_token")
	if err != ErrInvalidToken {
		t.Errorf("Expected error for invalid token: got %v", err)
	}
}

func TestValidateToken_Expired(t *testing.T) {
	userID := setupAuthTestDB(t)
	defer teardownAuthTestDB(t)

	// Create expired session
	token, _ := GenerateToken()
	expiredTime := time.Now().Add(-1 * time.Hour)
	database.DB.Exec(
		"INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, ?)",
		userID, token, expiredTime,
	)

	_, err := ValidateToken(token)
	if err != ErrExpiredToken {
		t.Errorf("Expected error for expired token: got %v", err)
	}
}

func TestDeleteSession(t *testing.T) {
	userID := setupAuthTestDB(t)
	defer teardownAuthTestDB(t)

	session, _ := CreateSession(userID)

	// Delete session
	err := DeleteSession(session.Token)
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Session should now be invalid
	_, err = ValidateToken(session.Token)
	if err != ErrInvalidToken {
		t.Error("Deleted session should still be valid")
	}
}

func TestCleanExpiredSessions(t *testing.T) {
	userID := setupAuthTestDB(t)
	defer teardownAuthTestDB(t)

	// Valid session
	CreateSession(userID)

	// Expired session
	token, _ := GenerateToken()
	expiredTime := time.Now().Add(-1 * time.Hour)
	database.DB.Exec(
		"INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, ?)",
		userID, token, expiredTime,
	)

	// Clean
	err := CleanExpiredSessions()
	if err != nil {
		t.Fatalf("Failed to clean expired sessions: %v", err)
	}

	// Expired session should be deleted
	_, err = ValidateToken(token)
	if err != ErrInvalidToken {
		t.Error("Expired session was not cleaned")
	}
}
