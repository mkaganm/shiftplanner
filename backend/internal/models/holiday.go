package models

import (
	"time"
)

// Turkey's public holidays (2025-2026)
var holidays = map[string]string{
	// 2025 - Official Holidays
	"2025-01-01": "New Year's Day",
	"2025-04-23": "National Sovereignty and Children's Day",
	"2025-05-01": "Labor and Solidarity Day",
	"2025-05-19": "Commemoration of Atatürk, Youth and Sports Day",
	"2025-07-15": "Democracy and National Unity Day",
	"2025-08-30": "Victory Day",
	"2025-10-29": "Republic Day",
	"2025-12-31": "New Year's Eve",

	// 2025 - Religious Holidays (Eid al-Fitr)
	"2025-03-29": "Eid al-Fitr Eve",
	"2025-03-30": "Eid al-Fitr",
	"2025-03-31": "Eid al-Fitr",
	"2025-04-01": "Eid al-Fitr",

	// 2025 - Religious Holidays (Eid al-Adha)
	"2025-06-05": "Eid al-Adha Eve",
	"2025-06-06": "Eid al-Adha",
	"2025-06-07": "Eid al-Adha",
	"2025-06-08": "Eid al-Adha",
	"2025-06-09": "Eid al-Adha",

	// 2026 - Official Holidays
	"2026-01-01": "New Year's Day",
	"2026-04-23": "National Sovereignty and Children's Day",
	"2026-05-01": "Labor and Solidarity Day",
	"2026-05-19": "Commemoration of Atatürk, Youth and Sports Day",
	"2026-07-15": "Democracy and National Unity Day",
	"2026-08-30": "Victory Day",
	"2026-10-29": "Republic Day",
	"2026-12-31": "New Year's Eve",

	// 2026 - Religious Holidays (Eid al-Fitr)
	"2026-03-18": "Eid al-Fitr Eve",
	"2026-03-19": "Eid al-Fitr",
	"2026-03-20": "Eid al-Fitr",
	"2026-03-21": "Eid al-Fitr",

	// 2026 - Religious Holidays (Eid al-Adha)
	"2026-05-25": "Eid al-Adha Eve",
	"2026-05-26": "Eid al-Adha",
	"2026-05-27": "Eid al-Adha",
	"2026-05-28": "Eid al-Adha",
	"2026-05-29": "Eid al-Adha",
}

// IsHoliday checks if the specified date is a public holiday
func IsHoliday(date time.Time) bool {
	dateStr := date.Format("2006-01-02")
	_, exists := holidays[dateStr]
	return exists
}

// IsWeekend checks if the specified date is a weekend
func IsWeekend(date time.Time) bool {
	weekday := date.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// IsWorkingDay checks if the specified date is a working day
func IsWorkingDay(date time.Time) bool {
	return !IsHoliday(date) && !IsWeekend(date)
}

// GetNextWorkingDay returns the first working day after the specified date
func GetNextWorkingDay(date time.Time) time.Time {
	nextDay := date.AddDate(0, 0, 1)
	for !IsWorkingDay(nextDay) {
		nextDay = nextDay.AddDate(0, 0, 1)
	}
	return nextDay
}

// GetPreviousWorkingDay returns the last working day before the specified date
func GetPreviousWorkingDay(date time.Time) time.Time {
	prevDay := date.AddDate(0, 0, -1)
	for !IsWorkingDay(prevDay) {
		prevDay = prevDay.AddDate(0, 0, -1)
	}
	return prevDay
}

// GetHolidayName returns the public holiday name for the specified date
func GetHolidayName(date time.Time) string {
	dateStr := date.Format("2006-01-02")
	return holidays[dateStr]
}

// GetAllHolidays returns all public holidays
func GetAllHolidays() map[string]string {
	return holidays
}

// WillBeLongShift checks if there is a holiday/weekend in the days following the specified date
// If so, this date will be the start of a long shift
func WillBeLongShift(date time.Time) bool {
	nextDay := date.AddDate(0, 0, 1)
	return IsHoliday(nextDay) || IsWeekend(nextDay)
}
