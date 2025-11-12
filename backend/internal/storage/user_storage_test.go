package storage

import (
	"database/sql"
	"os"
	"path/filepath"
	"shiftplanner/backend/internal/database"
	"testing"

	_ "modernc.org/sqlite"
)

func setupUserStorageTestDB(t *testing.T) {
	testDBPath := filepath.Join(os.TempDir(), "test_user_storage.db")
	os.Remove(testDBPath)

	var err error
	database.DB, err = sql.Open("sqlite", testDBPath)
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

	if _, err := database.DB.Exec(createUsersTable); err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}
}

func teardownUserStorageTestDB(t *testing.T) {
	if database.DB != nil {
		database.DB.Close()
		database.DB = nil
	}
	testDBPath := filepath.Join(os.TempDir(), "test_user_storage.db")
	os.Remove(testDBPath)
}

func TestCreateUser(t *testing.T) {
	setupUserStorageTestDB(t)
	defer teardownUserStorageTestDB(t)

	user, err := CreateUser("testuser", "testpassword")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.Username != "testuser" {
		t.Errorf("Username mismatch: got %s, want testuser", user.Username)
	}

	if user.ID == 0 {
		t.Error("User ID cannot be 0")
	}
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	setupUserStorageTestDB(t)
	defer teardownUserStorageTestDB(t)

	CreateUser("testuser", "password1")

	// Try to create again with same username
	_, err := CreateUser("testuser", "password2")
	if err == nil {
		t.Error("Expected error for duplicate username")
	}
}

func TestGetUserByUsername(t *testing.T) {
	setupUserStorageTestDB(t)
	defer teardownUserStorageTestDB(t)

	CreateUser("testuser", "testpassword")

	user, passwordHash, err := GetUserByUsername("testuser")
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if user.Username != "testuser" {
		t.Errorf("Username mismatch: got %s, want testuser", user.Username)
	}

	if passwordHash == "" {
		t.Error("Password hash cannot be empty")
	}
}

func TestGetUserByUsername_NotFound(t *testing.T) {
	setupUserStorageTestDB(t)
	defer teardownUserStorageTestDB(t)

	_, _, err := GetUserByUsername("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

func TestValidatePassword(t *testing.T) {
	setupUserStorageTestDB(t)
	defer teardownUserStorageTestDB(t)

	password := "testpassword"
	user, _ := CreateUser("testuser", password)

	// Get password hash
	_, passwordHash, _ := GetUserByUsername("testuser")

	// Correct password
	if !ValidatePassword(password, passwordHash) {
		t.Error("Correct password was not validated")
	}

	// Wrong password
	if ValidatePassword("wrongpassword", passwordHash) {
		t.Error("Wrong password should not be validated")
	}

	_ = user // Prevent unused variable warning
}

func TestHashPassword(t *testing.T) {
	password := "testpassword"
	hash1 := hashPassword(password)
	hash2 := hashPassword(password)

	// Same password should produce same hash
	if hash1 != hash2 {
		t.Error("Different hash produced for same password")
	}

	// Hash should not be empty
	if hash1 == "" {
		t.Error("Hash cannot be empty")
	}

	// Different password should produce different hash
	hash3 := hashPassword("differentpassword")
	if hash1 == hash3 {
		t.Error("Same hash produced for different passwords")
	}
}
