package api

import (
	"encoding/json"
	"net/http"
	"shiftplanner/backend/internal/models"
	"shiftplanner/backend/internal/scheduler"
	"shiftplanner/backend/internal/storage"
	"strconv"
	"strings"
	"time"
)

// GetMembers returns all members
func GetMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := GetUserID(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	members, err := storage.GetAllMembers(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

// CreateMember creates a new member
func CreateMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := GetUserID(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	member, err := storage.CreateMember(userID, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(member)
}

// DeleteMember deletes a member
func DeleteMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := GetUserID(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL: /api/members/:id
	path := strings.TrimPrefix(r.URL.Path, "/api/members/")
	memberID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	if err := storage.DeleteMember(userID, memberID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetShifts returns shifts
func GetShifts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			http.Error(w, "Invalid start_date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
	} else {
		startDate = time.Now().AddDate(0, -1, 0) // Last 1 month
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			http.Error(w, "Invalid end_date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
	} else {
		endDate = time.Now().AddDate(0, 1, 0) // Next 1 month
	}

	userID := GetUserID(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	shifts, err := storage.GetShiftsByDateRange(userID, startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Üye isimlerini ekle
	members, err := storage.GetAllMembers(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	memberMap := make(map[int]string)
	for _, m := range members {
		memberMap[m.ID] = m.Name
	}

	for i := range shifts {
		shifts[i].MemberName = memberMap[shifts[i].MemberID]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shifts)
}

// GenerateShifts creates a new shift plan
func GenerateShifts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req scheduler.PlanShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.StartDate.IsZero() || req.EndDate.IsZero() {
		http.Error(w, "start_date and end_date are required", http.StatusBadRequest)
		return
	}

	userID := GetUserID(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Create plan
	shifts, err := scheduler.PlanShift(userID, req.StartDate, req.EndDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete existing shifts (in the same date range)
	if err := storage.DeleteShiftsByDateRange(userID, req.StartDate, req.EndDate); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save new shifts
	for _, shift := range shifts {
		_, err := storage.CreateShift(userID, shift.MemberID, shift.StartDate, shift.EndDate, shift.IsLongShift)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Add member names
	members, err := storage.GetAllMembers(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	memberMap := make(map[int]string)
	for _, m := range members {
		memberMap[m.ID] = m.Name
	}

	for i := range shifts {
		shifts[i].MemberName = memberMap[shifts[i].MemberID]
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(shifts)
}

// GetHolidays returns public holidays
func GetHolidays(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	holidays := models.GetAllHolidays()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(holidays)
}

// GetStats returns member statistics
func GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := GetUserID(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	members, err := storage.GetAllMembers(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stats, err := storage.GetAllMembersStats(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
