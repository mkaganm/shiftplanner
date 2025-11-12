package storage

import (
	"database/sql"
	"os"
	"path/filepath"
	"shiftplanner/backend/internal/database"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) int {
	// Create temporary database for testing
	testDBPath := filepath.Join(os.TempDir(), "test_shifts.db")
	os.Remove(testDBPath) // Remove if exists

	var err error
	database.DB, err = sql.Open("sqlite3", testDBPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create test database schema
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

	if _, err := database.DB.Exec(createUsersTable); err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}
	if _, err := database.DB.Exec(createMembersTable); err != nil {
		t.Fatalf("Failed to create members table: %v", err)
	}
	if _, err := database.DB.Exec(createShiftsTable); err != nil {
		t.Fatalf("Failed to create shifts table: %v", err)
	}

	// Create test user
	result, _ := database.DB.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", "testuser", "testhash")
	userID, _ := result.LastInsertId()
	return int(userID)
}

func teardownTestDB(t *testing.T) {
	if database.DB != nil {
		database.DB.Close()
		database.DB = nil
	}
	// Clean up test database
	testDBPath := filepath.Join(os.TempDir(), "test_shifts.db")
	os.Remove(testDBPath)
}

func TestCreateMember(t *testing.T) {
	userID := setupTestDB(t)
	defer teardownTestDB(t)

	member, err := CreateMember(userID, "Test Member")
	if err != nil {
		t.Fatalf("Failed to create member: %v", err)
	}

	if member.Name != "Test Member" {
		t.Errorf("Member name mismatch: got %s, want Test Member", member.Name)
	}

	if member.ID == 0 {
		t.Error("Member ID cannot be 0")
	}
}

func TestGetAllMembers(t *testing.T) {
	userID := setupTestDB(t)
	defer teardownTestDB(t)

	// Add test data
	CreateMember(userID, "Member 1")
	CreateMember(userID, "Member 2")

	members, err := GetAllMembers(userID)
	if err != nil {
		t.Fatalf("Failed to get members: %v", err)
	}

	if len(members) < 2 {
		t.Errorf("Expected member count: at least 2, got %d", len(members))
	}
}

func TestDeleteMember(t *testing.T) {
	userID := setupTestDB(t)
	defer teardownTestDB(t)

	member, _ := CreateMember(userID, "Member to Delete")
	err := DeleteMember(userID, member.ID)
	if err != nil {
		t.Fatalf("Failed to delete member: %v", err)
	}

	members, _ := GetAllMembers(userID)
	for _, m := range members {
		if m.ID == member.ID {
			t.Error("Member was not deleted")
		}
	}
}

func TestCreateShift(t *testing.T) {
	userID := setupTestDB(t)
	defer teardownTestDB(t)

	member, _ := CreateMember(userID, "Test Member")
	startDate := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)

	shift, err := CreateShift(userID, member.ID, startDate, endDate, false)
	if err != nil {
		t.Fatalf("Failed to create shift: %v", err)
	}

	if shift.MemberID != member.ID {
		t.Errorf("Shift member ID mismatch: got %d, want %d", shift.MemberID, member.ID)
	}

	if shift.IsLongShift {
		t.Error("Shift should not be marked as long shift")
	}
}

func TestGetShiftsByDateRange(t *testing.T) {
	userID := setupTestDB(t)
	defer teardownTestDB(t)

	member, _ := CreateMember(userID, "Test Member")
	startDate := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)

	CreateShift(userID, member.ID, startDate, endDate, false)

	shifts, err := GetShiftsByDateRange(userID, startDate, endDate)
	if err != nil {
		t.Fatalf("Failed to get shifts: %v", err)
	}

	if len(shifts) == 0 {
		t.Error("No shift found")
	}
}

func TestGetMemberShiftStats(t *testing.T) {
	userID := setupTestDB(t)
	defer teardownTestDB(t)

	member, _ := CreateMember(userID, "Test Member")
	startDate := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC)

	CreateShift(userID, member.ID, startDate, endDate, true)

	totalDays, longShiftCount, err := GetMemberShiftStats(userID, member.ID)
	if err != nil {
		t.Fatalf("Failed to get statistics: %v", err)
	}

	if totalDays < 3 {
		t.Errorf("Total days count mismatch: got %d, want at least 3", totalDays)
	}

	if longShiftCount != 1 {
		t.Errorf("Long shift count mismatch: got %d, want 1", longShiftCount)
	}
}

func TestGetAllMembersStats(t *testing.T) {
	userID := setupTestDB(t)
	defer teardownTestDB(t)

	member1, _ := CreateMember(userID, "Member 1")
	member2, _ := CreateMember(userID, "Member 2")

	startDate1 := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	endDate1 := time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC)
	CreateShift(userID, member1.ID, startDate1, endDate1, true)

	startDate2 := time.Date(2025, 1, 9, 0, 0, 0, 0, time.UTC)
	endDate2 := time.Date(2025, 1, 9, 0, 0, 0, 0, time.UTC)
	CreateShift(userID, member2.ID, startDate2, endDate2, false)

	stats, err := GetAllMembersStats(userID)
	if err != nil {
		t.Fatalf("Failed to get statistics: %v", err)
	}

	if len(stats) != 2 {
		t.Errorf("Expected member count: 2, got %d", len(stats))
	}

	if stats[member1.ID].LongShiftCount != 1 {
		t.Errorf("Member 1's long shift count mismatch: got %d, want 1", stats[member1.ID].LongShiftCount)
	}

	if stats[member2.ID].LongShiftCount != 0 {
		t.Errorf("Member 2's long shift count mismatch: got %d, want 0", stats[member2.ID].LongShiftCount)
	}
}

func TestDeleteShiftsByDateRange(t *testing.T) {
	userID := setupTestDB(t)
	defer teardownTestDB(t)

	member, _ := CreateMember(userID, "Test Member")
	startDate := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)

	CreateShift(userID, member.ID, startDate, endDate, false)

	// Delete shifts in date range
	deleteStart := time.Date(2025, 1, 7, 0, 0, 0, 0, time.UTC)
	deleteEnd := time.Date(2025, 1, 9, 0, 0, 0, 0, time.UTC)
	err := DeleteShiftsByDateRange(userID, deleteStart, deleteEnd)
	if err != nil {
		t.Fatalf("Failed to delete shifts: %v", err)
	}

	// Check that shifts were deleted
	shifts, _ := GetShiftsByDateRange(userID, startDate, endDate)
	if len(shifts) > 0 {
		t.Error("Shifts were not deleted")
	}
}

func TestGetMemberByID(t *testing.T) {
	userID := setupTestDB(t)
	defer teardownTestDB(t)

	member, _ := CreateMember(userID, "Test Member")

	retrievedMember, err := GetMemberByID(userID, member.ID)
	if err != nil {
		t.Fatalf("Failed to get member: %v", err)
	}

	if retrievedMember.ID != member.ID {
		t.Errorf("Member ID mismatch: got %d, want %d", retrievedMember.ID, member.ID)
	}

	if retrievedMember.Name != "Test Member" {
		t.Errorf("Member name mismatch: got %s, want Test Member", retrievedMember.Name)
	}
}

func TestGetMemberByID_WrongUser(t *testing.T) {
	userID1 := setupTestDB(t)
	defer teardownTestDB(t)

	// Create second user
	result, _ := database.DB.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", "user2", "hash2")
	userID2, _ := result.LastInsertId()

	member, _ := CreateMember(userID1, "User1 Member")

	// User2 should not see User1's member
	_, err := GetMemberByID(int(userID2), member.ID)
	if err == nil {
		t.Error("Should not get another user's member")
	}
}
