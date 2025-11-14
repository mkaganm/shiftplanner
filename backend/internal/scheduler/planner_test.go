package scheduler

import (
	"testing"
	"time"
)

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
