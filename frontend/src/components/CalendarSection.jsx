// Calendar Section Component
import { useApp } from '../context/AppContext';
import './CalendarSection.css';

export function CalendarSection() {
  const { currentMonth, currentYear, shifts, holidays, previousMonth, nextMonth, formatDate } = useApp();

  const monthNames = [
    'January', 'February', 'March', 'April', 'May', 'June',
    'July', 'August', 'September', 'October', 'November', 'December'
  ];

  const weekDays = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];

  const firstDay = new Date(currentYear, currentMonth, 1);
  const lastDay = new Date(currentYear, currentMonth + 1, 0);
  const daysInMonth = lastDay.getDate();

  // getDay() returns 0=Sunday, 1=Monday, etc. We want Monday=0
  let startingDayOfWeek = firstDay.getDay() - 1;
  if (startingDayOfWeek < 0) startingDayOfWeek = 6; // Sunday becomes 6

  const renderDay = (day) => {
    if (day === 0) return null;

    const date = new Date(currentYear, currentMonth, day);
    const dateStr = formatDate(date);
    const dayOfWeek = date.getDay();

    // Check if it's a holiday
    const holidayName = holidays[dateStr];
    const isHoliday = !!holidayName;

    // Filter shifts for this day
    const dayShifts = shifts.filter((shift) => {
      if (!shift.start_date || !shift.end_date) return false;
      
      const startDateStr = shift.start_date.split('T')[0];
      const endDateStr = shift.end_date.split('T')[0];
      
      return dateStr >= startDateStr && dateStr <= endDateStr;
    });

    let classes = 'day';
    if (dayOfWeek === 0 || dayOfWeek === 6) {
      classes += ' weekend';
    }
    if (isHoliday) {
      classes += ' holiday';
    }

    return (
      <div key={day} className={classes}>
        <span className={isHoliday ? 'day-number holiday-day' : 'day-number'}>
          {day}
        </span>
        {isHoliday && (
          <span className="holiday-name" title={holidayName}>
            {holidayName}
          </span>
        )}
        {dayShifts.map((shift) => (
          <span
            key={shift.id}
            className={`shift-name ${shift.is_long_shift ? 'long-shift' : ''}`}
          >
            {shift.member_name || 'Unknown'}
          </span>
        ))}
      </div>
    );
  };

  return (
    <section className="section">
      <h2>Shift Plan</h2>
      <div className="calendar-controls">
        <button onClick={previousMonth}>← Previous Month</button>
        <span id="currentMonth">
          {monthNames[currentMonth]} {currentYear}
        </span>
        <button onClick={nextMonth}>Next Month →</button>
      </div>
      <div className="calendar">
        {weekDays.map((day) => (
          <div key={day} className="calendar-header">
            {day}
          </div>
        ))}
        {Array.from({ length: startingDayOfWeek }, (_, i) => (
          <div key={`empty-${i}`} className="day empty"></div>
        ))}
        {Array.from({ length: daysInMonth }, (_, i) => renderDay(i + 1))}
      </div>
    </section>
  );
}

