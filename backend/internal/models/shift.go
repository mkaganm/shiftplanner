package models

import (
	"time"
)

// Shift shift model
type Shift struct {
	ID          int       `json:"id"`
	MemberID    int       `json:"member_id"`
	MemberName  string    `json:"member_name,omitempty"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	IsLongShift bool      `json:"is_long_shift"`
	CreatedAt   time.Time `json:"created_at"`
}

// MemberStats member statistics
type MemberStats struct {
	TotalDays      int `json:"total_days"`
	LongShiftCount int `json:"long_shift_count"`
}

