package storage

import (
	"database/sql"
	"fmt"
	"log"
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

	// Normalize dates to UTC midnight
	startDateUTC := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	endDateUTC := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)

	startDateStr := startDateUTC.Format("2006-01-02")
	endDateStr := endDateUTC.Format("2006-01-02")

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
		StartDate:   startDateUTC,
		EndDate:     endDateUTC,
		IsLongShift: isLongShift,
		CreatedAt:   time.Now().UTC(),
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

		// Parse start_date - normalize to UTC midnight
		// Support both "2006-01-02" and ISO 8601 formats
		if startDateStr == "" {
			log.Printf("Warning: Shift ID %d has empty start_date, skipping", s.ID)
			continue
		}
		
		var startTime time.Time
		
		// Try parsing as "2006-01-02" first
		if t, parseErr := time.Parse("2006-01-02", startDateStr); parseErr == nil {
			startTime = t
		} else {
			// Try parsing as ISO 8601 format (with time)
			if t, parseErr := time.Parse(time.RFC3339, startDateStr); parseErr == nil {
				startTime = t
			} else {
				// Try parsing as "2006-01-02T15:04:05Z" format
				if t, parseErr := time.Parse("2006-01-02T15:04:05Z", startDateStr); parseErr == nil {
					startTime = t
				} else {
					log.Printf("Error parsing start_date '%s' for shift ID %d: %v", startDateStr, s.ID, parseErr)
					continue
				}
			}
		}
		
		// Normalize to UTC midnight
		s.StartDate = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC)

		// Parse end_date - normalize to UTC midnight
		// Support both "2006-01-02" and ISO 8601 formats
		if endDateStr == "" {
			log.Printf("Warning: Shift ID %d has empty end_date, skipping", s.ID)
			continue
		}
		
		var endTime time.Time
		
		// Try parsing as "2006-01-02" first
		if t, parseErr := time.Parse("2006-01-02", endDateStr); parseErr == nil {
			endTime = t
		} else {
			// Try parsing as ISO 8601 format (with time)
			if t, parseErr := time.Parse(time.RFC3339, endDateStr); parseErr == nil {
				endTime = t
			} else {
				// Try parsing as "2006-01-02T15:04:05Z" format
				if t, parseErr := time.Parse("2006-01-02T15:04:05Z", endDateStr); parseErr == nil {
					endTime = t
				} else {
					log.Printf("Error parsing end_date '%s' for shift ID %d: %v", endDateStr, s.ID, parseErr)
					continue
				}
			}
		}
		
		// Normalize to UTC midnight
		s.EndDate = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 0, 0, 0, 0, time.UTC)

		s.IsLongShift = isLongShift == 1
		// Parse SQLite datetime format
		if t, err := time.Parse("2006-01-02 15:04:05", createdAtStr); err == nil {
			s.CreatedAt = t.UTC()
		} else if t, err := time.Parse("2006-01-02T15:04:05Z07:00", createdAtStr); err == nil {
			s.CreatedAt = t.UTC()
		} else {
			s.CreatedAt = time.Now().UTC()
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

// DeleteShiftsByDateRange deletes shifts that overlap with the date range
// This ensures we don't have duplicate shifts when regenerating
func DeleteShiftsByDateRange(userID int, startDate, endDate time.Time) error {
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	// Delete shifts that overlap with the date range
	// A shift overlaps if: start_date <= endDate AND end_date >= startDate
	_, err := database.DB.Exec(
		"DELETE FROM shifts WHERE user_id = ? AND start_date <= ? AND end_date >= ?",
		userID, endDateStr, startDateStr,
	)
	return err
}

// DeleteAllShifts deletes all shifts for a user
func DeleteAllShifts(userID int) error {
	_, err := database.DB.Exec(
		"DELETE FROM shifts WHERE user_id = ?",
		userID,
	)
	return err
}

// GetShiftByDate gets a shift that covers a specific date
func GetShiftByDate(userID int, date time.Time) (*models.Shift, error) {
	dateStr := date.Format("2006-01-02")
	
	var s models.Shift
	var startDateStr, endDateStr, createdAtStr string
	var isLongShift int
	
	err := database.DB.QueryRow(
		"SELECT id, member_id, start_date, end_date, is_long_shift, created_at FROM shifts WHERE user_id = ? AND start_date <= ? AND end_date >= ? LIMIT 1",
		userID, dateStr, dateStr,
	).Scan(&s.ID, &s.MemberID, &startDateStr, &endDateStr, &isLongShift, &createdAtStr)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No shift found
		}
		return nil, err
	}

	// Parse start_date
	var startTime time.Time
	if t, parseErr := time.Parse("2006-01-02", startDateStr); parseErr == nil {
		startTime = t
	} else if t, parseErr := time.Parse(time.RFC3339, startDateStr); parseErr == nil {
		startTime = t
	} else if t, parseErr := time.Parse("2006-01-02T15:04:05Z", startDateStr); parseErr == nil {
		startTime = t
	} else {
		return nil, fmt.Errorf("error parsing start_date '%s': %v", startDateStr, parseErr)
	}
	s.StartDate = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC)

	// Parse end_date
	var endTime time.Time
	if t, parseErr := time.Parse("2006-01-02", endDateStr); parseErr == nil {
		endTime = t
	} else if t, parseErr := time.Parse(time.RFC3339, endDateStr); parseErr == nil {
		endTime = t
	} else if t, parseErr := time.Parse("2006-01-02T15:04:05Z", endDateStr); parseErr == nil {
		endTime = t
	} else {
		return nil, fmt.Errorf("error parsing end_date '%s': %v", endDateStr, parseErr)
	}
	s.EndDate = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 0, 0, 0, 0, time.UTC)

	s.IsLongShift = isLongShift == 1
	
	// Parse created_at
	if t, err := time.Parse("2006-01-02 15:04:05", createdAtStr); err == nil {
		s.CreatedAt = t.UTC()
	} else if t, err := time.Parse("2006-01-02T15:04:05Z07:00", createdAtStr); err == nil {
		s.CreatedAt = t.UTC()
	} else {
		s.CreatedAt = time.Now().UTC()
	}

	return &s, nil
}

// UpdateShiftMember updates the member for a shift
func UpdateShiftMember(userID, shiftID, newMemberID int) error {
	_, err := database.DB.Exec(
		"UPDATE shifts SET member_id = ? WHERE id = ? AND user_id = ?",
		newMemberID, shiftID, userID,
	)
	return err
}

// CreateOrUpdateShiftForDate creates or updates a shift for a specific date
// If a shift exists for that date, updates the member_id
// If no shift exists, creates a new single-day shift
func CreateOrUpdateShiftForDate(userID, memberID int, date time.Time) (*models.Shift, error) {
	// Normalize date to UTC midnight
	dateUTC := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	// Check if shift exists for this date
	existingShift, err := GetShiftByDate(userID, dateUTC)
	if err != nil {
		return nil, err
	}

	if existingShift != nil {
		// Update existing shift member
		if err := UpdateShiftMember(userID, existingShift.ID, memberID); err != nil {
			return nil, err
		}
		
		// Fetch updated shift
		updatedShift, err := GetShiftByDate(userID, dateUTC)
		if err != nil {
			return nil, err
		}
		return updatedShift, nil
	}

	// Create new single-day shift
	// Check if it should be a long shift
	isLongShift := false
	nextDay := dateUTC.AddDate(0, 0, 1)
	if models.IsHoliday(nextDay) || models.IsWeekend(nextDay) {
		isLongShift = true
	}

	return CreateShift(userID, memberID, dateUTC, dateUTC, isLongShift)
}

// CreateLeaveDay creates a new leave day record
func CreateLeaveDay(userID, memberID int, leaveDate time.Time) (*models.LeaveDay, error) {
	// Validate date is not zero
	if leaveDate.IsZero() {
		return nil, fmt.Errorf("leave_date cannot be zero")
	}

	// Normalize date to UTC midnight
	leaveDateUTC := time.Date(leaveDate.Year(), leaveDate.Month(), leaveDate.Day(), 0, 0, 0, 0, time.UTC)
	leaveDateStr := leaveDateUTC.Format("2006-01-02")

	result, err := database.DB.Exec(
		"INSERT INTO leave_days (user_id, member_id, leave_date) VALUES (?, ?, ?)",
		userID, memberID, leaveDateStr,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.LeaveDay{
		ID:        int(id),
		MemberID:  memberID,
		LeaveDate: leaveDateUTC,
		CreatedAt: time.Now().UTC(),
	}, nil
}

// CreateLeaveDaysRange creates leave days for a date range
func CreateLeaveDaysRange(userID, memberID int, startDate, endDate time.Time) ([]models.LeaveDay, error) {
	if startDate.IsZero() || endDate.IsZero() {
		return nil, fmt.Errorf("start_date and end_date cannot be zero")
	}

	if startDate.After(endDate) {
		return nil, fmt.Errorf("start_date must be before or equal to end_date")
	}

	var leaveDays []models.LeaveDay
	currentDate := startDate

	// Iterate through each day in the range
	for !currentDate.After(endDate) {
		// Normalize to UTC midnight
		dateUTC := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, time.UTC)
		dateStr := dateUTC.Format("2006-01-02")

		// Check if leave day already exists
		var existingID int
		err := database.DB.QueryRow(
			"SELECT id FROM leave_days WHERE user_id = ? AND member_id = ? AND leave_date = ?",
			userID, memberID, dateStr,
		).Scan(&existingID)

		if err == nil {
			// Already exists, fetch it
			var existingLeaveDay models.LeaveDay
			var createdAtStr string
			err := database.DB.QueryRow(
				"SELECT id, member_id, leave_date, created_at FROM leave_days WHERE id = ?",
				existingID,
			).Scan(&existingLeaveDay.ID, &existingLeaveDay.MemberID, &dateStr, &createdAtStr)
			if err == nil {
				// Parse the date
				if t, parseErr := time.Parse("2006-01-02", dateStr); parseErr == nil {
					existingLeaveDay.LeaveDate = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
				}
				// Parse created_at
				if t, err := time.Parse("2006-01-02 15:04:05", createdAtStr); err == nil {
					existingLeaveDay.CreatedAt = t.UTC()
				} else if t, err := time.Parse("2006-01-02T15:04:05Z07:00", createdAtStr); err == nil {
					existingLeaveDay.CreatedAt = t.UTC()
				} else {
					existingLeaveDay.CreatedAt = time.Now().UTC()
				}
				leaveDays = append(leaveDays, existingLeaveDay)
			}
		} else {
			// Doesn't exist, insert it
			result, err := database.DB.Exec(
				"INSERT INTO leave_days (user_id, member_id, leave_date) VALUES (?, ?, ?)",
				userID, memberID, dateStr,
			)
			if err != nil {
				return nil, err
			}

			id, err := result.LastInsertId()
			if err == nil {
				leaveDays = append(leaveDays, models.LeaveDay{
					ID:        int(id),
					MemberID:  memberID,
					LeaveDate: dateUTC,
					CreatedAt: time.Now().UTC(),
				})
			}
		}

		// Move to next day
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return leaveDays, nil
}

// GetLeaveDaysByDateRange gets leave days for members in a date range
func GetLeaveDaysByDateRange(userID int, startDate, endDate time.Time) ([]models.LeaveDay, error) {
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	rows, err := database.DB.Query(
		"SELECT id, member_id, leave_date, created_at FROM leave_days WHERE user_id = ? AND leave_date >= ? AND leave_date <= ? ORDER BY leave_date",
		userID, startDateStr, endDateStr,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaveDays []models.LeaveDay
	for rows.Next() {
		var ld models.LeaveDay
		var leaveDateStr, createdAtStr string

		if err := rows.Scan(&ld.ID, &ld.MemberID, &leaveDateStr, &createdAtStr); err != nil {
			return nil, err
		}

		// Parse leave_date - normalize to UTC midnight
		if leaveDateStr == "" {
			log.Printf("Warning: Leave day ID %d has empty leave_date, skipping", ld.ID)
			continue
		}

		var leaveTime time.Time
		if t, parseErr := time.Parse("2006-01-02", leaveDateStr); parseErr == nil {
			leaveTime = t
		} else if t, parseErr := time.Parse(time.RFC3339, leaveDateStr); parseErr == nil {
			leaveTime = t
		} else if t, parseErr := time.Parse("2006-01-02T15:04:05Z", leaveDateStr); parseErr == nil {
			leaveTime = t
		} else {
			log.Printf("Error parsing leave_date '%s' for leave day ID %d: %v", leaveDateStr, ld.ID, parseErr)
			continue
		}

		// Normalize to UTC midnight
		ld.LeaveDate = time.Date(leaveTime.Year(), leaveTime.Month(), leaveTime.Day(), 0, 0, 0, 0, time.UTC)

		// Parse created_at
		if t, err := time.Parse("2006-01-02 15:04:05", createdAtStr); err == nil {
			ld.CreatedAt = t.UTC()
		} else if t, err := time.Parse("2006-01-02T15:04:05Z07:00", createdAtStr); err == nil {
			ld.CreatedAt = t.UTC()
		} else {
			ld.CreatedAt = time.Now().UTC()
		}

		leaveDays = append(leaveDays, ld)
	}

	return leaveDays, rows.Err()
}

// GetLeaveDaysByMember gets all leave days for a specific member
func GetLeaveDaysByMember(userID, memberID int) ([]models.LeaveDay, error) {
	rows, err := database.DB.Query(
		"SELECT id, member_id, leave_date, created_at FROM leave_days WHERE user_id = ? AND member_id = ? ORDER BY leave_date",
		userID, memberID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaveDays []models.LeaveDay
	for rows.Next() {
		var ld models.LeaveDay
		var leaveDateStr, createdAtStr string

		if err := rows.Scan(&ld.ID, &ld.MemberID, &leaveDateStr, &createdAtStr); err != nil {
			return nil, err
		}

		// Parse leave_date - normalize to UTC midnight
		if leaveDateStr == "" {
			log.Printf("Warning: Leave day ID %d has empty leave_date, skipping", ld.ID)
			continue
		}

		var leaveTime time.Time
		if t, parseErr := time.Parse("2006-01-02", leaveDateStr); parseErr == nil {
			leaveTime = t
		} else if t, parseErr := time.Parse(time.RFC3339, leaveDateStr); parseErr == nil {
			leaveTime = t
		} else if t, parseErr := time.Parse("2006-01-02T15:04:05Z", leaveDateStr); parseErr == nil {
			leaveTime = t
		} else {
			log.Printf("Error parsing leave_date '%s' for leave day ID %d: %v", leaveDateStr, ld.ID, parseErr)
			continue
		}

		// Normalize to UTC midnight
		ld.LeaveDate = time.Date(leaveTime.Year(), leaveTime.Month(), leaveTime.Day(), 0, 0, 0, 0, time.UTC)

		// Parse created_at
		if t, err := time.Parse("2006-01-02 15:04:05", createdAtStr); err == nil {
			ld.CreatedAt = t.UTC()
		} else if t, err := time.Parse("2006-01-02T15:04:05Z07:00", createdAtStr); err == nil {
			ld.CreatedAt = t.UTC()
		} else {
			ld.CreatedAt = time.Now().UTC()
		}

		leaveDays = append(leaveDays, ld)
	}

	return leaveDays, rows.Err()
}

// IsMemberOnLeave checks if a member is on leave on a specific date
func IsMemberOnLeave(userID, memberID int, date time.Time) (bool, error) {
	dateStr := date.Format("2006-01-02")
	var count int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM leave_days WHERE user_id = ? AND member_id = ? AND leave_date = ?",
		userID, memberID, dateStr,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// DeleteLeaveDay deletes a leave day record
func DeleteLeaveDay(userID, leaveDayID int) error {
	_, err := database.DB.Exec(
		"DELETE FROM leave_days WHERE id = ? AND user_id = ?",
		leaveDayID, userID,
	)
	return err
}
