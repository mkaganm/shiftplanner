package api

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"shiftplanner/backend/internal/auth"
	"shiftplanner/backend/internal/database"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupMiddlewareTestDB(t *testing.T) int {
	testDBPath := filepath.Join(os.TempDir(), "test_middleware.db")
	os.Remove(testDBPath)

	var err error
	database.DB, err = sql.Open("sqlite3", testDBPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

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

func teardownMiddlewareTestDB(t *testing.T) {
	if database.DB != nil {
		database.DB.Close()
		database.DB = nil
	}
	testDBPath := filepath.Join(os.TempDir(), "test_middleware.db")
	os.Remove(testDBPath)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	userID := setupMiddlewareTestDB(t)
	defer teardownMiddlewareTestDB(t)

	// Create valid session
	session, _ := auth.CreateSession(userID)

	// Test handler
	handlerCalled := false
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		// Check userID from context
		ctxUserID := GetUserID(r)
		if ctxUserID != userID {
			t.Errorf("User ID in context mismatch: got %d, want %d", ctxUserID, userID)
		}
		w.WriteHeader(http.StatusOK)
	}

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", session.Token)
	w := httptest.NewRecorder()

	// Call handler with middleware
	middleware := AuthMiddleware(testHandler)
	middleware(w, req)

	if !handlerCalled {
		t.Error("Handler was not called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Status code mismatch: got %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	setupMiddlewareTestDB(t)
	defer teardownMiddlewareTestDB(t)

	handlerCalled := false
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	middleware := AuthMiddleware(testHandler)
	middleware(w, req)

	if handlerCalled {
		t.Error("Handler should not be called")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status code mismatch: got %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	setupMiddlewareTestDB(t)
	defer teardownMiddlewareTestDB(t)

	handlerCalled := false
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "invalid_token")
	w := httptest.NewRecorder()

	middleware := AuthMiddleware(testHandler)
	middleware(w, req)

	if handlerCalled {
		t.Error("Handler should not be called")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status code mismatch: got %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestGetUserID(t *testing.T) {
	userID := setupMiddlewareTestDB(t)
	defer teardownMiddlewareTestDB(t)

	// Add userID to context
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), userIDContextKey, userID)
	req = req.WithContext(ctx)

	// Test GetUserID
	result := GetUserID(req)
	if result != userID {
		t.Errorf("User ID mismatch: got %d, want %d", result, userID)
	}
}

func TestGetUserID_NoContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	result := GetUserID(req)
	if result != 0 {
		t.Errorf("User ID should be 0: got %d", result)
	}
}
