package storage

import (
	"database/sql"
	"fmt"
	"shiftplanner/backend/internal/database"
	"shiftplanner/backend/internal/models"
	"time"
)

// GetAllMembers gets all members for a user
func GetAllMembers(userID int) ([]models.Member, error) {
	rows, err := database.DB.Query("SELECT id, name, created_at FROM members WHERE user_id = ? ORDER BY name", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.Member
	for rows.Next() {
		var m models.Member
		var createdAtStr string
		if err := rows.Scan(&m.ID, &m.Name, &createdAtStr); err != nil {
			return nil, err
		}
		// Parse SQLite datetime format
		if t, err := time.Parse("2006-01-02 15:04:05", createdAtStr); err == nil {
			m.CreatedAt = t
		} else if t, err := time.Parse("2006-01-02T15:04:05Z07:00", createdAtStr); err == nil {
			m.CreatedAt = t
		} else {
			m.CreatedAt = time.Now()
		}
		members = append(members, m)
	}

	return members, rows.Err()
}

// CreateMember creates a new member
func CreateMember(userID int, name string) (*models.Member, error) {
	result, err := database.DB.Exec("INSERT INTO members (user_id, name) VALUES (?, ?)", userID, name)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.Member{
		ID:        int(id),
		Name:      name,
		CreatedAt: time.Now(),
	}, nil
}

// DeleteMember deletes a member (can only delete own members)
func DeleteMember(userID, memberID int) error {
	_, err := database.DB.Exec("DELETE FROM members WHERE id = ? AND user_id = ?", memberID, userID)
	return err
}

// GetMemberByID gets a member by ID (can only get own members)
func GetMemberByID(userID, memberID int) (*models.Member, error) {
	var m models.Member
	var createdAtStr string
	err := database.DB.QueryRow("SELECT id, name, created_at FROM members WHERE id = ? AND user_id = ?", memberID, userID).
		Scan(&m.ID, &m.Name, &createdAtStr)
	if err != nil {
		return nil, err
	}
	// Parse SQLite datetime format
	if t, err := time.Parse("2006-01-02 15:04:05", createdAtStr); err == nil {
		m.CreatedAt = t
	} else if t, err := time.Parse("2006-01-02T15:04:05Z07:00", createdAtStr); err == nil {
		m.CreatedAt = t
	} else {
		m.CreatedAt = time.Now()
	}
	return &m, nil
}

// CreateShift creates a new shift record
func CreateShift(userID, memberID int, startDate, endDate time.Time, isLongShift bool) (*models.Shift, error) {
	// Validate dates are not zero
	if startDate.IsZero() || endDate.IsZero() {
		return nil, fmt.Errorf("start_date and end_date cannot be zero")
	}

	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	result, err := database.DB.Exec(
		"INSERT INTO shifts (user_id, member_id, start_date, end_date, is_long_shift) VALUES (?, ?, ?, ?, ?)",
		userID, memberID, startDateStr, endDateStr, isLongShift,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.Shift{
		ID:          int(id),
		MemberID:    memberID,
		StartDate:   startDate,
		EndDate:     endDate,
		IsLongShift: isLongShift,
		CreatedAt:   time.Now(),
	}, nil
}

// GetShiftsByDateRange gets shifts by date range
func GetShiftsByDateRange(userID int, startDate, endDate time.Time) ([]models.Shift, error) {
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	rows, err := database.DB.Query(
		"SELECT id, member_id, start_date, end_date, is_long_shift, created_at FROM shifts WHERE user_id = ? AND start_date <= ? AND end_date >= ? ORDER BY start_date",
		userID, endDateStr, startDateStr,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []models.Shift
	for rows.Next() {
		var s models.Shift
		var startDateStr, endDateStr, createdAtStr string
		var isLongShift int

		if err := rows.Scan(&s.ID, &s.MemberID, &startDateStr, &endDateStr, &isLongShift, &createdAtStr); err != nil {
			return nil, err
		}

		// Parse start_date - log error if parsing fails
		if startDateStr == "" {
			// Skip shifts with empty dates
			continue
		}
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			s.StartDate = t
		} else {
			// Log parse error but continue
			// This should not happen if data is correct
			continue
		}

		// Parse end_date - log error if parsing fails
		if endDateStr == "" {
			// Skip shifts with empty dates
			continue
		}
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			s.EndDate = t
		} else {
			// Log parse error but continue
			continue
		}

		s.IsLongShift = isLongShift == 1
		// Parse SQLite datetime format
		if t, err := time.Parse("2006-01-02 15:04:05", createdAtStr); err == nil {
			s.CreatedAt = t
		} else if t, err := time.Parse("2006-01-02T15:04:05Z07:00", createdAtStr); err == nil {
			s.CreatedAt = t
		} else {
			s.CreatedAt = time.Now()
		}

		shifts = append(shifts, s)
	}

	return shifts, rows.Err()
}

// GetMemberShiftStats gets shift statistics for a member
func GetMemberShiftStats(userID, memberID int) (totalDays int, longShiftCount int, err error) {
	// Total shift days count
	var totalDaysResult sql.NullInt64
	err = database.DB.QueryRow(`
		SELECT COALESCE(SUM(julianday(end_date) - julianday(start_date) + 1), 0) 
		FROM shifts 
		WHERE user_id = ? AND member_id = ?
	`, userID, memberID).Scan(&totalDaysResult)
	if err != nil {
		return 0, 0, err
	}
	if totalDaysResult.Valid {
		totalDays = int(totalDaysResult.Int64)
	}

	// Long shift count
	err = database.DB.QueryRow(`
		SELECT COUNT(*) 
		FROM shifts 
		WHERE user_id = ? AND member_id = ? AND is_long_shift = 1
	`, userID, memberID).Scan(&longShiftCount)
	if err != nil {
		return 0, 0, err
	}

	return totalDays, longShiftCount, nil
}

// GetAllMembersStats gets statistics for all members
func GetAllMembersStats(userID int) (map[int]models.MemberStats, error) {
	members, err := GetAllMembers(userID)
	if err != nil {
		return nil, err
	}

	stats := make(map[int]models.MemberStats)
	for _, member := range members {
		totalDays, longShiftCount, err := GetMemberShiftStats(userID, member.ID)
		if err != nil {
			return nil, err
		}
		stats[member.ID] = models.MemberStats{
			TotalDays:      totalDays,
			LongShiftCount: longShiftCount,
		}
	}

	return stats, nil
}

// DeleteShiftsByDateRange deletes shifts in a date range
func DeleteShiftsByDateRange(userID int, startDate, endDate time.Time) error {
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	_, err := database.DB.Exec(
		"DELETE FROM shifts WHERE user_id = ? AND start_date >= ? AND start_date <= ?",
		userID, startDateStr, endDateStr,
	)
	return err
}
