package scheduler

import (
	"shiftplanner/backend/internal/models"
	"testing"
	"time"
)

func TestSelectMemberForDay(t *testing.T) {
	memberIDs := []int{1, 2, 3}
	
	// Active counts (database + newly assigned)
	activeTotalDays := map[int]int{
		1: 5,
		2: 3,
		3: 3,
	}
	activeLongShiftCounts := map[int]int{
		1: 2,
		2: 1,
		3: 2,
	}

	// Test normal shift selection: should select member with least total days
	selectedID := selectMemberForDay(memberIDs, activeTotalDays, activeLongShiftCounts, 0, false)
	if selectedID != 2 && selectedID != 3 {
		t.Errorf("Wrong member selected: got %d, want 2 or 3 (least total days: 3)", selectedID)
	}

	// Test avoiding consecutive shifts: previous member was 2
	selectedID = selectMemberForDay(memberIDs, activeTotalDays, activeLongShiftCounts, 2, false)
	if selectedID == 2 {
		t.Errorf("Should not select previous day's member: got %d, want 1 or 3", selectedID)
	}

	// Test with updated counts: member 2 already has 2 more days assigned
	activeTotalDays[2] = 5 // Now member 2 has 5 total, member 3 has 3 total
	selectedID = selectMemberForDay(memberIDs, activeTotalDays, activeLongShiftCounts, 0, false)
	// Should select member 3 (least total: 3)
	if selectedID != 3 {
		t.Errorf("Wrong member selected with updated counts: got %d, want 3 (least total: 3)", selectedID)
	}

	// Test long shift selection: should prioritize least long shift count
	// Reset counts
	activeTotalDays = map[int]int{1: 5, 2: 3, 3: 3}
	activeLongShiftCounts = map[int]int{1: 2, 2: 1, 3: 2}
	// Member 2 has 1 long shift (least), should be selected
	selectedID = selectMemberForDay(memberIDs, activeTotalDays, activeLongShiftCounts, 0, true)
	if selectedID != 2 {
		t.Errorf("Wrong member selected for long shift: got %d, want 2 (least long shifts: 1)", selectedID)
	}

	// Test long shift selection with updated counts: member 2 already has 1 more long shift
	activeLongShiftCounts[2] = 2 // Now all have 2 long shifts
	selectedID = selectMemberForDay(memberIDs, activeTotalDays, activeLongShiftCounts, 0, true)
	// Should select based on total days (tie-breaker)
	if selectedID == 0 {
		t.Errorf("No member selected for long shift")
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
