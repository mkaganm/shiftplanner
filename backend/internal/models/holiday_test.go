package models

import (
	"testing"
	"time"
)

func TestIsHoliday(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected bool
	}{
		{
			name:     "New Year's Day 2025",
			date:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "April 23, 2025",
			date:     time.Date(2025, 4, 23, 0, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "Normal day",
			date:     time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsHoliday(tt.date)
			if result != tt.expected {
				t.Errorf("IsHoliday(%v) = %v, want %v", tt.date, result, tt.expected)
			}
		})
	}
}

func TestIsWeekend(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected bool
	}{
		{
			name:     "Saturday",
			date:     time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "Sunday",
			date:     time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "Monday",
			date:     time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWeekend(tt.date)
			if result != tt.expected {
				t.Errorf("IsWeekend(%v) = %v, want %v", tt.date, result, tt.expected)
			}
		})
	}
}

func TestIsWorkingDay(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected bool
	}{
		{
			name:     "Working day (Monday)",
			date:     time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC), // Monday
			expected: true,
		},
		{
			name:     "Weekend",
			date:     time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "Public holiday",
			date:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWorkingDay(tt.date)
			if result != tt.expected {
				t.Errorf("IsWorkingDay(%v) = %v, want %v", tt.date, result, tt.expected)
			}
		})
	}
}

func TestGetNextWorkingDay(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected time.Time
	}{
		{
			name:     "Working day after Saturday",
			date:     time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC), // Saturday
			expected: time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC), // Monday
		},
		{
			name:     "Working day after Sunday",
			date:     time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC), // Sunday
			expected: time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC), // Monday
		},
		{
			name:     "Working day after working day",
			date:     time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC), // Monday
			expected: time.Date(2025, 1, 7, 0, 0, 0, 0, time.UTC), // Tuesday
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetNextWorkingDay(tt.date)
			if !result.Equal(tt.expected) {
				t.Errorf("GetNextWorkingDay(%v) = %v, want %v", tt.date, result, tt.expected)
			}
		})
	}
}

func TestWillBeLongShift(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected bool
	}{
		{
			name:     "Friday (tomorrow is Saturday)",
			date:     time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC), // Friday
			expected: true,
		},
		{
			name:     "Thursday (tomorrow is Friday, working day)",
			date:     time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), // Thursday
			expected: false,
		},
		{
			name:     "Monday (tomorrow is Tuesday, working day)",
			date:     time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC), // Monday
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WillBeLongShift(tt.date)
			if result != tt.expected {
				t.Errorf("WillBeLongShift(%v) = %v, want %v", tt.date, result, tt.expected)
			}
		})
	}
}
