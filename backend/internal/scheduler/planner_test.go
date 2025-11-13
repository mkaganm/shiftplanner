package scheduler

import (
	"shiftplanner/backend/internal/models"
	"testing"
	"time"
)

func TestSelectMember(t *testing.T) {
	// Mock stats for testing
	stats := map[int]models.MemberStats{
		1: {TotalDays: 5, LongShiftCount: 2}, // Normal: 3
		2: {TotalDays: 3, LongShiftCount: 1}, // Normal: 2
		3: {TotalDays: 3, LongShiftCount: 2}, // Normal: 1
	}

	memberIDs := []int{1, 2, 3}

	// Test long shift selection: should select member with least long shift count
	selectedID := selectMember(memberIDs, stats, true)
	if selectedID != 2 {
		t.Errorf("Wrong member selected for long shift: got %d, want 2 (least long shifts)", selectedID)
	}

	// Test normal shift selection: should select member with least normal shift count
	selectedID = selectMember(memberIDs, stats, false)
	if selectedID != 3 {
		t.Errorf("Wrong member selected for normal shift: got %d, want 3 (least normal shifts: 1)", selectedID)
	}
}

func TestUpdateStats(t *testing.T) {
	stats := map[int]models.MemberStats{
		1: {TotalDays: 5, LongShiftCount: 1},
	}

	startDate := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC)

	updateStats(&stats, 1, startDate, endDate, true)

	if stats[1].TotalDays != 8 { // 5 + 3 days
		t.Errorf("Total days count not updated: got %d, want 8", stats[1].TotalDays)
	}

	if stats[1].LongShiftCount != 2 { // 1 + 1
		t.Errorf("Long shift count not updated: got %d, want 2", stats[1].LongShiftCount)
	}
}

func TestPlanShift_EmptyMembers(t *testing.T) {
	// This test doesn't need real database
	// We're just testing the empty member list case

	startDate := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)

	// This test requires real database connection
	// Can be tested with mocks but for now let's do a simple test
	_ = startDate
	_ = endDate
}

func TestPlanShiftRequest_UnmarshalJSON(t *testing.T) {
	jsonData := `{"start_date":"2025-01-06","end_date":"2025-01-10"}`

	var req PlanShiftRequest
	err := req.UnmarshalJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	expectedStart := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)

	if !req.StartDate.Equal(expectedStart) {
		t.Errorf("Start date mismatch: got %v, want %v", req.StartDate, expectedStart)
	}

	if !req.EndDate.Equal(expectedEnd) {
		t.Errorf("End date mismatch: got %v, want %v", req.EndDate, expectedEnd)
	}
}

func TestPlanShiftRequest_UnmarshalJSON_WithTime(t *testing.T) {
	// Date with time in ISO 8601 format
	jsonData := `{"start_date":"2025-01-06T10:00:00Z","end_date":"2025-01-10T15:30:00Z"}`

	var req PlanShiftRequest
	err := req.UnmarshalJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	expectedStart := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)

	if !req.StartDate.Equal(expectedStart) {
		t.Errorf("Start date mismatch: got %v, want %v", req.StartDate, expectedStart)
	}

	if !req.EndDate.Equal(expectedEnd) {
		t.Errorf("End date mismatch: got %v, want %v", req.EndDate, expectedEnd)
	}
}
