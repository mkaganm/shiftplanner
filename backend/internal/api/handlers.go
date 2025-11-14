package api

import (
	"log"
	"shiftplanner/backend/internal/models"
	"shiftplanner/backend/internal/scheduler"
	"shiftplanner/backend/internal/storage"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GetMembers returns all members
func GetMembers(c *fiber.Ctx) error {
	userID := GetUserID(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	members, err := storage.GetAllMembers(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(members)
}

// CreateMember creates a new member
func CreateMember(c *fiber.Ctx) error {
	userID := GetUserID(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name is required",
		})
	}

	member, err := storage.CreateMember(userID, req.Name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(member)
}

// DeleteMember deletes a member
func DeleteMember(c *fiber.Ctx) error {
	userID := GetUserID(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Extract ID from URL parameter
	memberID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid member ID",
		})
	}

	if err := storage.DeleteMember(userID, memberID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetShifts returns shifts
func GetShifts(c *fiber.Ctx) error {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid start_date format (use YYYY-MM-DD)",
			})
		}
		// Normalize to UTC midnight
		startDate = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, time.UTC)
	} else {
		now := time.Now().UTC()
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, -1, 0) // Last 1 month
	}

	if endDateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid end_date format (use YYYY-MM-DD)",
			})
		}
		// Normalize to UTC midnight
		endDate = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, time.UTC)
	} else {
		now := time.Now().UTC()
		endDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0) // Next 1 month
	}

	userID := GetUserID(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	shifts, err := storage.GetShiftsByDateRange(userID, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Add member names
	members, err := storage.GetAllMembers(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	memberMap := make(map[int]string)
	for _, m := range members {
		memberMap[m.ID] = m.Name
	}

	for i := range shifts {
		shifts[i].MemberName = memberMap[shifts[i].MemberID]
	}

	// Log shifts for debugging
	if len(shifts) > 0 {
		log.Printf("Returning %d shifts", len(shifts))
	} else {
		log.Printf("No shifts found for user %d in range %s to %s", userID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	}

	return c.JSON(shifts)
}

// GenerateShifts creates a new shift plan
func GenerateShifts(c *fiber.Ctx) error {
	var req scheduler.PlanShiftRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if req.StartDate.IsZero() || req.EndDate.IsZero() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "start_date and end_date are required",
		})
	}

	userID := GetUserID(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Create plan
	shifts, err := scheduler.PlanShift(userID, req.StartDate, req.EndDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Delete existing shifts (in the same date range)
	if err := storage.DeleteShiftsByDateRange(userID, req.StartDate, req.EndDate); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Save new shifts
	for _, shift := range shifts {
		_, err := storage.CreateShift(userID, shift.MemberID, shift.StartDate, shift.EndDate, shift.IsLongShift)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	// Add member names
	members, err := storage.GetAllMembers(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	memberMap := make(map[int]string)
	for _, m := range members {
		memberMap[m.ID] = m.Name
	}

	for i := range shifts {
		shifts[i].MemberName = memberMap[shifts[i].MemberID]
	}

	return c.Status(fiber.StatusCreated).JSON(shifts)
}

// GetHolidays returns public holidays
func GetHolidays(c *fiber.Ctx) error {
	holidays := models.GetAllHolidays()
	return c.JSON(holidays)
}

// GetStats returns member statistics
func GetStats(c *fiber.Ctx) error {
	userID := GetUserID(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	members, err := storage.GetAllMembers(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	stats, err := storage.GetAllMembersStats(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	type MemberStatResponse struct {
		MemberID       int    `json:"member_id"`
		MemberName     string `json:"member_name"`
		TotalDays      int    `json:"total_days"`
		LongShiftCount int    `json:"long_shift_count"`
	}

	var response []MemberStatResponse
	for _, member := range members {
		memberStats := stats[member.ID]
		response = append(response, MemberStatResponse{
			MemberID:       member.ID,
			MemberName:     member.Name,
			TotalDays:      memberStats.TotalDays,
			LongShiftCount: memberStats.LongShiftCount,
		})
	}

	return c.JSON(response)
}

// ClearAllShifts deletes all shifts for the authenticated user
func ClearAllShifts(c *fiber.Ctx) error {
	userID := GetUserID(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	if err := storage.DeleteAllShifts(userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// CreateLeaveDay creates leave days for a date range
func CreateLeaveDay(c *fiber.Ctx) error {
	userID := GetUserID(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req struct {
		MemberID  int    `json:"member_id"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if req.MemberID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "member_id is required",
		})
	}

	if req.StartDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "start_date is required",
		})
	}

	if req.EndDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "end_date is required",
		})
	}

	// Parse dates
	parsedStartDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid start_date format (use YYYY-MM-DD)",
		})
	}

	parsedEndDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid end_date format (use YYYY-MM-DD)",
		})
	}

	// Normalize to UTC midnight
	startDate := time.Date(parsedStartDate.Year(), parsedStartDate.Month(), parsedStartDate.Day(), 0, 0, 0, 0, time.UTC)
	endDate := time.Date(parsedEndDate.Year(), parsedEndDate.Month(), parsedEndDate.Day(), 0, 0, 0, 0, time.UTC)

	if startDate.After(endDate) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "start_date must be before or equal to end_date",
		})
	}

	leaveDays, err := storage.CreateLeaveDaysRange(userID, req.MemberID, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Add member name
	member, err := storage.GetMemberByID(userID, req.MemberID)
	if err == nil {
		for i := range leaveDays {
			leaveDays[i].MemberName = member.Name
		}
	}

	return c.Status(fiber.StatusCreated).JSON(leaveDays)
}

// GetLeaveDays returns leave days
func GetLeaveDays(c *fiber.Ctx) error {
	userID := GetUserID(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	memberIDStr := c.Query("member_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var leaveDays []models.LeaveDay
	var err error

	if memberIDStr != "" {
		// Get leave days for specific member
		memberID, err := strconv.Atoi(memberIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid member_id",
			})
		}
		leaveDays, err = storage.GetLeaveDaysByMember(userID, memberID)
	} else if startDateStr != "" && endDateStr != "" {
		// Get leave days for date range
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid start_date format (use YYYY-MM-DD)",
			})
		}
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid end_date format (use YYYY-MM-DD)",
			})
		}
		// Normalize to UTC midnight
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)
		leaveDays, err = storage.GetLeaveDaysByDateRange(userID, startDate, endDate)
	} else {
		// Get all leave days (last year to next year)
		now := time.Now().UTC()
		startDate := time.Date(now.Year()-1, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(now.Year()+1, 12, 31, 0, 0, 0, 0, time.UTC)
		leaveDays, err = storage.GetLeaveDaysByDateRange(userID, startDate, endDate)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Add member names
	members, err := storage.GetAllMembers(userID)
	if err == nil {
		memberMap := make(map[int]string)
		for _, m := range members {
			memberMap[m.ID] = m.Name
		}
		for i := range leaveDays {
			leaveDays[i].MemberName = memberMap[leaveDays[i].MemberID]
		}
	}

	return c.JSON(leaveDays)
}

// DeleteLeaveDay deletes a leave day
func DeleteLeaveDay(c *fiber.Ctx) error {
	userID := GetUserID(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	leaveDayID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid leave day ID",
		})
	}

	if err := storage.DeleteLeaveDay(userID, leaveDayID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateShiftForDate updates or creates a shift for a specific date
func UpdateShiftForDate(c *fiber.Ctx) error {
	userID := GetUserID(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req struct {
		Date     string `json:"date"`
		MemberID int    `json:"member_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if req.Date == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "date is required",
		})
	}

	if req.MemberID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "member_id is required",
		})
	}

	// Parse date
	parsedDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid date format (use YYYY-MM-DD)",
		})
	}

	// Normalize to UTC midnight
	date := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, time.UTC)

	// Create or update shift
	shift, err := storage.CreateOrUpdateShiftForDate(userID, req.MemberID, date)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Add member name
	member, err := storage.GetMemberByID(userID, req.MemberID)
	if err == nil {
		shift.MemberName = member.Name
	}

	return c.JSON(shift)
}
