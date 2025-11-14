package scheduler

import (
	"encoding/json"
	"math/rand"
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

	// Convert member IDs to a slice
	memberIDs := make([]int, len(members))
	for i, m := range members {
		memberIDs[i] = m.ID
	}

	// Two separate maps: normal shift days and long shift days
	// Key: memberID, Value: shift days count
	normalShiftDays := make(map[int]int) // memberID -> total normal shift days
	longShiftDays := make(map[int]int)   // memberID -> total long shift days

	// Get all shifts from database and calculate actual day counts
	allShifts, err := storage.GetShiftsByDateRange(userID, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC))
	if err == nil {
		// Calculate days for each shift and add to corresponding map
		for _, shift := range allShifts {
			days := int(shift.EndDate.Sub(shift.StartDate).Hours()/24) + 1
			if shift.IsLongShift {
				longShiftDays[shift.MemberID] += days
			} else {
				normalShiftDays[shift.MemberID] += days
			}
		}
	}

	// Initialize maps for all members (0 for members without shifts)
	for _, id := range memberIDs {
		if _, exists := normalShiftDays[id]; !exists {
			normalShiftDays[id] = 0
		}
		if _, exists := longShiftDays[id]; !exists {
			longShiftDays[id] = 0
		}
	}

	// Get existing shifts (for conflict check)
	existingShifts, err := storage.GetShiftsByDateRange(userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Delete conflicting shifts (to overwrite)
	if len(existingShifts) > 0 {
		if err := storage.DeleteShiftsByDateRange(userID, startDate, endDate); err != nil {
			return nil, err
		}
	}

	// Get leave days for the planning period
	leaveDays, err := storage.GetLeaveDaysByDateRange(userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Create a map of member IDs to leave dates for quick lookup
	// Key: date string (YYYY-MM-DD), Value: set of member IDs on leave
	memberLeaveMap := make(map[string]map[int]bool)
	for _, ld := range leaveDays {
		dateStr := ld.LeaveDate.Format("2006-01-02")
		if memberLeaveMap[dateStr] == nil {
			memberLeaveMap[dateStr] = make(map[int]bool)
		}
		memberLeaveMap[dateStr][ld.MemberID] = true
	}

	// Track which member was on duty for each day
	// Key: date string (YYYY-MM-DD), Value: memberID
	prevDayMemberMap := make(map[string]int)

	var shifts []models.Shift

	// Iterate through each day we want to assign shifts
	currentDate := startDate
	for !currentDate.After(endDate) {
		// Only process working days
		if !models.IsWorkingDay(currentDate) {
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}

		// Find member who was on duty on previous working day
		// If there's a weekend or holiday in between, use the last working day with shift
		prevWorkingDay := models.GetPreviousWorkingDay(currentDate)
		prevDateStr := prevWorkingDay.Format("2006-01-02")
		prevDayMemberID := prevDayMemberMap[prevDateStr]

		// Get members on leave for this date
		currentDateStr := currentDate.Format("2006-01-02")
		membersOnLeave := memberLeaveMap[currentDateStr]

		// Is this day a long shift?
		isLongShift := models.WillBeLongShift(currentDate)

		// Select appropriate member
		var selectedMemberID int
		if isLongShift {
			// Long shift: select member with least long shift days
			selectedMemberID = selectMemberByLongShift(memberIDs, longShiftDays, prevDayMemberID, membersOnLeave)
		} else {
			// Normal shift: select member with least normal shift days
			selectedMemberID = selectMemberByNormalShift(memberIDs, normalShiftDays, prevDayMemberID, membersOnLeave)
		}

		// Calculate shift end date
		endDateForShift := currentDate
		if isLongShift {
			// Long shift continues until next working day
			nextWorkingDay := models.GetNextWorkingDay(currentDate)
			endDateForShift = nextWorkingDay.AddDate(0, 0, -1)
		}
		if endDateForShift.After(endDate) {
			endDateForShift = endDate
		}

		// Create new shift
		shift := models.Shift{
			MemberID:    selectedMemberID,
			StartDate:   currentDate,
			EndDate:     endDateForShift,
			IsLongShift: isLongShift,
			CreatedAt:   time.Now(),
		}
		shifts = append(shifts, shift)

		// Update shift day counts immediately (for next day)
		shiftDays := int(endDateForShift.Sub(currentDate).Hours()/24) + 1
		if isLongShift {
			longShiftDays[selectedMemberID] += shiftDays
		} else {
			normalShiftDays[selectedMemberID] += shiftDays
		}

		// Save member who was on duty today (for next day)
		prevDayMemberMap[currentDateStr] = selectedMemberID

		// Move to next day
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return shifts, nil
}

// selectMemberByNormalShift selects member with least normal shift days
// Excludes member who was on duty previous day (prevents consecutive shifts)
// Excludes members on leave for the current date
// Makes random selection if there's a tie
func selectMemberByNormalShift(memberIDs []int, normalShiftDays map[int]int, prevMemberID int, membersOnLeave map[int]bool) int {
	if len(memberIDs) == 0 {
		return 0
	}

	// Exclude member who was on duty previous day and members on leave
	availableIDs := make([]int, 0)
	for _, id := range memberIDs {
		if id != prevMemberID && !membersOnLeave[id] {
			availableIDs = append(availableIDs, id)
		}
	}

	// If all members were on duty yesterday or on leave, select one anyway (excluding those on leave)
	if len(availableIDs) == 0 {
		for _, id := range memberIDs {
			if !membersOnLeave[id] {
				availableIDs = append(availableIDs, id)
			}
		}
		// If all members are on leave, return 0 (no assignment possible)
		if len(availableIDs) == 0 {
			return 0
		}
	}

	// Find members with least normal shift days
	minNormalDays := normalShiftDays[availableIDs[0]]
	for _, id := range availableIDs {
		if normalShiftDays[id] < minNormalDays {
			minNormalDays = normalShiftDays[id]
		}
	}

	// Collect all members with minimum value
	candidates := make([]int, 0)
	for _, id := range availableIDs {
		if normalShiftDays[id] == minNormalDays {
			candidates = append(candidates, id)
		}
	}

	// Make random selection
	if len(candidates) > 0 {
		return candidates[rand.Intn(len(candidates))]
	}

	return availableIDs[0]
}

// selectMemberByLongShift selects member with least long shift days
// Excludes member who was on duty previous day (prevents consecutive shifts)
// Excludes members on leave for the current date
// Makes random selection if there's a tie
func selectMemberByLongShift(memberIDs []int, longShiftDays map[int]int, prevMemberID int, membersOnLeave map[int]bool) int {
	if len(memberIDs) == 0 {
		return 0
	}

	// Exclude member who was on duty previous day and members on leave
	availableIDs := make([]int, 0)
	for _, id := range memberIDs {
		if id != prevMemberID && !membersOnLeave[id] {
			availableIDs = append(availableIDs, id)
		}
	}

	// If all members were on duty yesterday or on leave, select one anyway (excluding those on leave)
	if len(availableIDs) == 0 {
		for _, id := range memberIDs {
			if !membersOnLeave[id] {
				availableIDs = append(availableIDs, id)
			}
		}
		// If all members are on leave, return 0 (no assignment possible)
		if len(availableIDs) == 0 {
			return 0
		}
	}

	// Find members with least long shift days
	minLongDays := longShiftDays[availableIDs[0]]
	for _, id := range availableIDs {
		if longShiftDays[id] < minLongDays {
			minLongDays = longShiftDays[id]
		}
	}

	// Collect all members with minimum value
	candidates := make([]int, 0)
	for _, id := range availableIDs {
		if longShiftDays[id] == minLongDays {
			candidates = append(candidates, id)
		}
	}

	// Make random selection
	if len(candidates) > 0 {
		return candidates[rand.Intn(len(candidates))]
	}

	return availableIDs[0]
}
