package scheduler

import (
	"encoding/json"
	"shiftplanner/backend/internal/models"
	"shiftplanner/backend/internal/storage"
	"strings"
	"time"
)

// PlanShiftRequest planning request
type PlanShiftRequest struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// UnmarshalJSON custom JSON unmarshaler - supports "YYYY-MM-DD" format
func (p *PlanShiftRequest) UnmarshalJSON(data []byte) error {
	var aux struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Normalize date format (take only date part)
	startDateStr := strings.Split(aux.StartDate, "T")[0]
	endDateStr := strings.Split(aux.EndDate, "T")[0]

	var err error
	p.StartDate, err = time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return err
	}

	p.EndDate, err = time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return err
	}

	return nil
}

// PlanShift creates a shift plan for the specified date range
func PlanShift(userID int, startDate, endDate time.Time) ([]models.Shift, error) {
	// Get existing members
	members, err := storage.GetAllMembers(userID)
	if err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return []models.Shift{}, nil
	}

	// Get existing statistics
	stats, err := storage.GetAllMembersStats(userID)
	if err != nil {
		return nil, err
	}

	// Convert member IDs to a slice
	memberIDs := make([]int, len(members))
	for i, m := range members {
		memberIDs[i] = m.ID
	}

	var shifts []models.Shift
	currentShiftMemberID := 0
	currentDate := startDate

	// If start date is not a working day, go to first working day
	if !models.IsWorkingDay(currentDate) {
		currentDate = models.GetNextWorkingDay(currentDate)
	}

	// Loop through date range
	for !currentDate.After(endDate) {
		// Is today a working day?
		if models.IsWorkingDay(currentDate) {
			// Will it be a long shift? (is tomorrow a holiday or weekend?)
			isLongShift := models.WillBeLongShift(currentDate)

			// Select member based on shift type
			// Long shift: select member with least long shift count
			// Normal shift: select member with least normal shift count
			selectedMemberID := selectMember(memberIDs, stats, isLongShift)

			// Calculate shift end date
			endDateForShift := currentDate
			if isLongShift {
				// Continues until next working day
				nextWorkingDay := models.GetNextWorkingDay(currentDate)
				// Take previous day (last day of holiday/weekend)
				endDateForShift = nextWorkingDay.AddDate(0, 0, -1)
			}

			// End date should not exceed endDate
			if endDateForShift.After(endDate) {
				endDateForShift = endDate
			}

			// Create shift record
			shift := models.Shift{
				MemberID:    selectedMemberID,
				StartDate:   currentDate,
				EndDate:     endDateForShift,
				IsLongShift: isLongShift,
				CreatedAt:   time.Now(),
			}

			// Update statistics
			updateStats(&stats, selectedMemberID, shift.StartDate, shift.EndDate, isLongShift)

			shifts = append(shifts, shift)
			currentShiftMemberID = selectedMemberID

			// Advance date until shift ends
			currentDate = endDateForShift.AddDate(0, 0, 1)
		} else {
			// Holiday or weekend - previous shift person continues
			// If there's no previous shift person (shouldn't happen as we checked at start), skip to first working day
			if currentShiftMemberID == 0 {
				currentDate = models.GetNextWorkingDay(currentDate)
			} else {
				// Previous shift person continues, advance date
				currentDate = currentDate.AddDate(0, 0, 1)
			}
		}
	}

	return shifts, nil
}

// selectMember selects the most suitable member
// For long shifts: selects member with least long shift count
// For normal shifts: selects member with least normal shift count (total days - long shift days)
func selectMember(memberIDs []int, stats map[int]models.MemberStats, isLongShift bool) int {
	if len(memberIDs) == 0 {
		return 0
	}

	selectedID := memberIDs[0]

	if isLongShift {
		// For long shifts: select member with least long shift count
		minLongShifts := stats[selectedID].LongShiftCount

		for _, id := range memberIDs {
			memberStats := stats[id]
			if memberStats.LongShiftCount < minLongShifts {
				selectedID = id
				minLongShifts = memberStats.LongShiftCount
			} else if memberStats.LongShiftCount == minLongShifts {
				// In case of tie, select member with least total days
				if memberStats.TotalDays < stats[selectedID].TotalDays {
					selectedID = id
				}
			}
		}
	} else {
		// For normal shifts: select member with least normal shift count
		// Normal shift count = total days - (long shift count * average long shift days)
		// For simplicity, we use total days - long shift count as approximation
		// Better approach: calculate actual normal shift days
		selectedID = memberIDs[0]
		minNormalShifts := calculateNormalShiftCount(stats[selectedID])

		for _, id := range memberIDs {
			memberStats := stats[id]
			normalShiftCount := calculateNormalShiftCount(memberStats)

			if normalShiftCount < minNormalShifts {
				selectedID = id
				minNormalShifts = normalShiftCount
			} else if normalShiftCount == minNormalShifts {
				// In case of tie, select member with least total days
				if memberStats.TotalDays < stats[selectedID].TotalDays {
					selectedID = id
				}
			}
		}
	}

	return selectedID
}

// calculateNormalShiftCount calculates the number of normal shift days
// Normal shift days = total days - long shift days
// We approximate long shift days as long shift count * average days per long shift
// For simplicity, we use: total days - long shift count
// This gives us a reasonable approximation for normal shift distribution
func calculateNormalShiftCount(stats models.MemberStats) int {
	// Normal shift count approximation: total days - long shift count
	// This assumes each long shift is roughly 1 day more than normal shift
	// In reality, we should calculate actual normal shift days from database
	return stats.TotalDays - stats.LongShiftCount
}

// updateStats updates statistics
func updateStats(stats *map[int]models.MemberStats, memberID int, startDate, endDate time.Time, isLongShift bool) {
	memberStats := (*stats)[memberID]

	// Calculate total days count
	days := int(endDate.Sub(startDate).Hours()/24) + 1
	memberStats.TotalDays += days

	// Increment long shift count
	if isLongShift {
		memberStats.LongShiftCount++
	}

	(*stats)[memberID] = memberStats
}
