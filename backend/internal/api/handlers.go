package api

import (
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
