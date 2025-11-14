package models

import (
	"time"
)

// LeaveDay leave day model
type LeaveDay struct {
	ID        int       `json:"id"`
	MemberID  int       `json:"member_id"`
	MemberName string   `json:"member_name,omitempty"`
	LeaveDate time.Time `json:"leave_date"`
	CreatedAt time.Time `json:"created_at"`
}

