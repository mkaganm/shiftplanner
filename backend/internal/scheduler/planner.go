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

	// İki ayrı map: normal shift gün sayısı ve long shift gün sayısı
	// Key: memberID, Value: shift gün sayısı
	normalShiftDays := make(map[int]int) // memberID -> normal shift toplam gün sayısı
	longShiftDays := make(map[int]int)   // memberID -> long shift toplam gün sayısı

	// Veritabanındaki tüm shift'leri al ve gerçek gün sayılarını hesapla
	allShifts, err := storage.GetShiftsByDateRange(userID, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC))
	if err == nil {
		// Her shift için gün sayısını hesapla ve ilgili map'e ekle
		for _, shift := range allShifts {
			days := int(shift.EndDate.Sub(shift.StartDate).Hours()/24) + 1
			if shift.IsLongShift {
				longShiftDays[shift.MemberID] += days
			} else {
				normalShiftDays[shift.MemberID] += days
			}
		}
	}

	// Tüm üyeler için map'leri başlat (shift'i olmayanlar için 0)
	for _, id := range memberIDs {
		if _, exists := normalShiftDays[id]; !exists {
			normalShiftDays[id] = 0
		}
		if _, exists := longShiftDays[id]; !exists {
			longShiftDays[id] = 0
		}
	}

	// Mevcut shift'leri al (çakışma kontrolü için)
	existingShifts, err := storage.GetShiftsByDateRange(userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Çakışan shift'leri sil (üzerine yazmak için)
	if len(existingShifts) > 0 {
		if err := storage.DeleteShiftsByDateRange(userID, startDate, endDate); err != nil {
			return nil, err
		}
	}

	// Her gün için bir önceki gün nöbet tutan kişiyi takip et
	// Key: tarih string (YYYY-MM-DD), Value: memberID
	prevDayMemberMap := make(map[string]int)

	var shifts []models.Shift

	// Nöbet atamak istediğimiz günleri tek tek gez
	currentDate := startDate
	for !currentDate.After(endDate) {
		// Sadece çalışma günlerini işle
		if !models.IsWorkingDay(currentDate) {
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}

		// Bir önceki çalışma günü nöbet tutan kişiyi bul
		// Eğer araya hafta sonu veya tatil girerse, son nöbet tutulan çalışma günü baz alınır
		prevWorkingDay := models.GetPreviousWorkingDay(currentDate)
		prevDateStr := prevWorkingDay.Format("2006-01-02")
		prevDayMemberID := prevDayMemberMap[prevDateStr]

		// Bu gün long shift mi?
		isLongShift := models.WillBeLongShift(currentDate)

		// Uygun kişiyi seç
		var selectedMemberID int
		if isLongShift {
			// Long shift: en az long shift gününe sahip kişiyi seç
			selectedMemberID = selectMemberByLongShift(memberIDs, longShiftDays, prevDayMemberID)
		} else {
			// Normal shift: en az normal shift gününe sahip kişiyi seç
			selectedMemberID = selectMemberByNormalShift(memberIDs, normalShiftDays, prevDayMemberID)
		}

		// Shift bitiş tarihini hesapla
		endDateForShift := currentDate
		if isLongShift {
			// Long shift bir sonraki çalışma gününe kadar devam eder
			nextWorkingDay := models.GetNextWorkingDay(currentDate)
			endDateForShift = nextWorkingDay.AddDate(0, 0, -1)
		}
		if endDateForShift.After(endDate) {
			endDateForShift = endDate
		}

		// Yeni shift oluştur
		shift := models.Shift{
			MemberID:    selectedMemberID,
			StartDate:   currentDate,
			EndDate:     endDateForShift,
			IsLongShift: isLongShift,
			CreatedAt:   time.Now(),
		}
		shifts = append(shifts, shift)

		// Shift gün sayılarını anında güncelle (bir sonraki gün için)
		shiftDays := int(endDateForShift.Sub(currentDate).Hours()/24) + 1
		if isLongShift {
			longShiftDays[selectedMemberID] += shiftDays
		} else {
			normalShiftDays[selectedMemberID] += shiftDays
		}

		// Bu gün nöbet tutan kişiyi kaydet (bir sonraki gün için)
		currentDateStr := currentDate.Format("2006-01-02")
		prevDayMemberMap[currentDateStr] = selectedMemberID

		// Bir sonraki güne geç
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return shifts, nil
}

// selectMemberByNormalShift en az normal shift gününe sahip kişiyi seçer
// Önceki gün nöbet tutan kişiyi hariç tutar (arka arkaya nöbet engelleme)
// Eşitlik durumunda rastgele seçim yapar
func selectMemberByNormalShift(memberIDs []int, normalShiftDays map[int]int, prevMemberID int) int {
	if len(memberIDs) == 0 {
		return 0
	}

	// Önceki gün nöbet tutan kişiyi hariç tut
	availableIDs := make([]int, 0)
	for _, id := range memberIDs {
		if id != prevMemberID {
			availableIDs = append(availableIDs, id)
		}
	}

	// Eğer tüm kişiler dün nöbet tuttuysa, yine de birini seç
	if len(availableIDs) == 0 {
		availableIDs = memberIDs
	}

	// En az normal shift gününe sahip kişileri bul
	minNormalDays := normalShiftDays[availableIDs[0]]
	for _, id := range availableIDs {
		if normalShiftDays[id] < minNormalDays {
			minNormalDays = normalShiftDays[id]
		}
	}

	// En az değere sahip tüm kişileri topla
	candidates := make([]int, 0)
	for _, id := range availableIDs {
		if normalShiftDays[id] == minNormalDays {
			candidates = append(candidates, id)
		}
	}

	// Rastgele seçim yap
	if len(candidates) > 0 {
		return candidates[rand.Intn(len(candidates))]
	}

	return availableIDs[0]
}

// selectMemberByLongShift en az long shift gününe sahip kişiyi seçer
// Önceki gün nöbet tutan kişiyi hariç tutar (arka arkaya nöbet engelleme)
// Eşitlik durumunda rastgele seçim yapar
func selectMemberByLongShift(memberIDs []int, longShiftDays map[int]int, prevMemberID int) int {
	if len(memberIDs) == 0 {
		return 0
	}

	// Önceki gün nöbet tutan kişiyi hariç tut
	availableIDs := make([]int, 0)
	for _, id := range memberIDs {
		if id != prevMemberID {
			availableIDs = append(availableIDs, id)
		}
	}

	// Eğer tüm kişiler dün nöbet tuttuysa, yine de birini seç
	if len(availableIDs) == 0 {
		availableIDs = memberIDs
	}

	// En az long shift gününe sahip kişileri bul
	minLongDays := longShiftDays[availableIDs[0]]
	for _, id := range availableIDs {
		if longShiftDays[id] < minLongDays {
			minLongDays = longShiftDays[id]
		}
	}

	// En az değere sahip tüm kişileri topla
	candidates := make([]int, 0)
	for _, id := range availableIDs {
		if longShiftDays[id] == minLongDays {
			candidates = append(candidates, id)
		}
	}

	// Rastgele seçim yap
	if len(candidates) > 0 {
		return candidates[rand.Intn(len(candidates))]
	}

	return availableIDs[0]
}
