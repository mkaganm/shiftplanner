package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"shiftplanner/backend/internal/database"
	"shiftplanner/backend/internal/storage"
	"strconv"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	_ "modernc.org/sqlite"
)

func setupTestAPI(t *testing.T) int {
	return setupTestDB(t)
}

func setupTestDB(t *testing.T) int {
	testDBPath := "test_api.db"
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

	createMembersTable := `
	CREATE TABLE IF NOT EXISTS members (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	createShiftsTable := `
	CREATE TABLE IF NOT EXISTS shifts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		member_id INTEGER NOT NULL,
		start_date DATE NOT NULL,
		end_date DATE NOT NULL,
		is_long_shift BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE
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
	if _, err := database.DB.Exec(createMembersTable); err != nil {
		t.Fatalf("Failed to create members table: %v", err)
	}
	if _, err := database.DB.Exec(createShiftsTable); err != nil {
		t.Fatalf("Failed to create shifts table: %v", err)
	}
	if _, err := database.DB.Exec(createSessionsTable); err != nil {
		t.Fatalf("Failed to create sessions table: %v", err)
	}

	result, _ := database.DB.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", "testuser", "testhash")
	userID, _ := result.LastInsertId()
	return int(userID)
}

func teardownTestAPI(t *testing.T) {
	if database.DB != nil {
		database.DB.Close()
		database.DB = nil
	}
	os.Remove("test_api.db")
}

func TestGetShifts_NoDatabase(t *testing.T) {
	database.DB = nil

	app := fiber.New()
	app.Get("/api/shifts", GetShifts)

	req := httptest.NewRequest(http.MethodGet, "/api/shifts?start_date=2025-01-06&end_date=2025-01-06", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode == http.StatusOK {
		t.Error("Should not succeed without database connection")
	}
}

func TestGetMembers(t *testing.T) {
	userID := setupTestDB(t)
	defer teardownTestAPI(t)

	storage.CreateMember(userID, "Test Member 1")
	storage.CreateMember(userID, "Test Member 2")

	token := "test_token_123"
	database.DB.Exec("INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, datetime('now', '+7 days'))", userID, token)

	app := fiber.New()
	app.Get("/api/members", AuthMiddleware, GetMembers)

	req := httptest.NewRequest(http.MethodGet, "/api/members", nil)
	req.Header.Set("Authorization", token)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code: %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var members []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	json.NewDecoder(resp.Body).Decode(&members)

	if len(members) < 2 {
		t.Errorf("Expected member count: at least 2, got %d", len(members))
	}
}

func TestCreateMember(t *testing.T) {
	userID := setupTestDB(t)
	defer teardownTestAPI(t)

	token := "test_token_123"
	database.DB.Exec("INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, datetime('now', '+7 days'))", userID, token)

	app := fiber.New()
	app.Post("/api/members", AuthMiddleware, CreateMember)

	body := bytes.NewBufferString(`{"name":"New Member"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/members", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code: %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	var member struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	json.NewDecoder(resp.Body).Decode(&member)

	if member.Name != "New Member" {
		t.Errorf("Member name mismatch: got %s, want New Member", member.Name)
	}
}

func TestCreateMember_EmptyName(t *testing.T) {
	userID := setupTestDB(t)
	defer teardownTestAPI(t)

	token := "test_token_123"
	database.DB.Exec("INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, datetime('now', '+7 days'))", userID, token)

	app := fiber.New()
	app.Post("/api/members", AuthMiddleware, CreateMember)

	body := bytes.NewBufferString(`{"name":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/members", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code: %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestGetShifts(t *testing.T) {
	userID := setupTestDB(t)
	defer teardownTestAPI(t)

	token := "test_token_123"
	database.DB.Exec("INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, datetime('now', '+7 days'))", userID, token)

	member, err := storage.CreateMember(userID, "Test Member")
	if err != nil {
		t.Fatalf("Failed to create member: %v", err)
	}

	startDate := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	_, err = storage.CreateShift(userID, member.ID, startDate, endDate, false)
	if err != nil {
		t.Fatalf("Failed to create shift: %v", err)
	}

	app := fiber.New()
	app.Get("/api/shifts", AuthMiddleware, GetShifts)

	req := httptest.NewRequest(http.MethodGet, "/api/shifts?start_date=2025-01-06&end_date=2025-01-06", nil)
	req.Header.Set("Authorization", token)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code: %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestGetStats(t *testing.T) {
	userID := setupTestDB(t)
	defer teardownTestAPI(t)

	token := "test_token_123"
	database.DB.Exec("INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, datetime('now', '+7 days'))", userID, token)

	storage.CreateMember(userID, "Test Member 1")
	storage.CreateMember(userID, "Test Member 2")

	app := fiber.New()
	app.Get("/api/stats", AuthMiddleware, GetStats)

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	req.Header.Set("Authorization", token)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code: %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestDeleteMember(t *testing.T) {
	userID := setupTestDB(t)
	defer teardownTestAPI(t)

	token := "test_token_123"
	database.DB.Exec("INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, datetime('now', '+7 days'))", userID, token)

	member, _ := storage.CreateMember(userID, "Member to Delete")

	app := fiber.New()
	app.Delete("/api/members/:id", AuthMiddleware, DeleteMember)

	req := httptest.NewRequest(http.MethodDelete, "/api/members/"+strconv.Itoa(member.ID), nil)
	req.Header.Set("Authorization", token)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status code: %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
}

func TestDeleteMember_InvalidID(t *testing.T) {
	userID := setupTestDB(t)
	defer teardownTestAPI(t)

	token := "test_token_123"
	database.DB.Exec("INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, datetime('now', '+7 days'))", userID, token)

	app := fiber.New()
	app.Delete("/api/members/:id", AuthMiddleware, DeleteMember)

	req := httptest.NewRequest(http.MethodDelete, "/api/members/invalid", nil)
	req.Header.Set("Authorization", token)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code: %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestGenerateShifts(t *testing.T) {
	userID := setupTestDB(t)
	defer teardownTestAPI(t)

	token := "test_token_123"
	database.DB.Exec("INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, datetime('now', '+7 days'))", userID, token)

	storage.CreateMember(userID, "Member 1")
	storage.CreateMember(userID, "Member 2")

	app := fiber.New()
	app.Post("/api/shifts/generate", AuthMiddleware, GenerateShifts)

	body := bytes.NewBufferString(`{"start_date":"2025-01-06","end_date":"2025-01-10"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/shifts/generate", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code: %d, got %d", http.StatusCreated, resp.StatusCode)
	}
}

func TestGetHolidays(t *testing.T) {
	app := fiber.New()
	app.Get("/api/holidays", GetHolidays)

	req := httptest.NewRequest(http.MethodGet, "/api/holidays", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code: %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var holidays []struct {
		Date string `json:"date"`
		Name string `json:"name"`
	}
	json.NewDecoder(resp.Body).Decode(&holidays)

	if len(holidays) == 0 {
		t.Error("Holiday list should not be empty")
	}
}

func TestRegister(t *testing.T) {
	setupTestDB(t)
	defer teardownTestAPI(t)

	app := fiber.New()
	app.Post("/api/auth/register", Register)

	body := bytes.NewBufferString(`{"username":"newuser","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code: %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	if response["token"] == nil {
		t.Error("Token should be returned")
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	setupTestDB(t)
	defer teardownTestAPI(t)

	database.DB.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", "existinguser", "hash")

	app := fiber.New()
	app.Post("/api/auth/register", Register)

	body := bytes.NewBufferString(`{"username":"existinguser","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Expected status code: %d, got %d", http.StatusConflict, resp.StatusCode)
	}
}

func TestLogin(t *testing.T) {
	setupTestDB(t)
	defer teardownTestAPI(t)

	storage.CreateUser("testuser", "password123")

	app := fiber.New()
	app.Post("/api/auth/login", Login)

	body := bytes.NewBufferString(`{"username":"testuser","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code: %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	if response["token"] == nil {
		t.Error("Token should be returned")
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	setupTestDB(t)
	defer teardownTestAPI(t)

	app := fiber.New()
	app.Post("/api/auth/login", Login)

	body := bytes.NewBufferString(`{"username":"nonexistent","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status code: %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestLogout(t *testing.T) {
	userID := setupTestDB(t)
	defer teardownTestAPI(t)

	token := "test_token_123"
	database.DB.Exec("INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, datetime('now', '+7 days'))", userID, token)

	app := fiber.New()
	app.Post("/api/auth/logout", AuthMiddleware, Logout)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.Header.Set("Authorization", token)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code: %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
