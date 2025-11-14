// Calendar Section Component
import { useState } from 'react';
import { useApp } from '../context/AppContext';
import './CalendarSection.css';

export function CalendarSection() {
  const { currentMonth, currentYear, shifts, holidays, leaveDays, members, previousMonth, nextMonth, formatDate, updateShiftForDate, loading } = useApp();
  const [showShiftModal, setShowShiftModal] = useState(null);
  const [selectedMemberId, setSelectedMemberId] = useState('');
  const [shiftError, setShiftError] = useState('');

  // Debug: Log shifts
  console.log('CalendarSection - shifts:', shifts);

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
      if (!shift || !shift.start_date || !shift.end_date) {
        return false;
      }
      
      // Handle both string and Date object formats
      let startDateStr, endDateStr;
      
      if (typeof shift.start_date === 'string') {
        startDateStr = shift.start_date.split('T')[0];
      } else if (shift.start_date instanceof Date) {
        startDateStr = formatDate(shift.start_date);
      } else {
        return false;
      }
      
      if (typeof shift.end_date === 'string') {
        endDateStr = shift.end_date.split('T')[0];
      } else if (shift.end_date instanceof Date) {
        endDateStr = formatDate(shift.end_date);
      } else {
        return false;
      }
      
      return dateStr >= startDateStr && dateStr <= endDateStr;
    });

    // Filter leave days for this day
    const dayLeaveDays = leaveDays.filter((leaveDay) => {
      if (!leaveDay || !leaveDay.leave_date) {
        return false;
      }
      
      // Handle both string and Date object formats
      let leaveDateStr;
      
      if (typeof leaveDay.leave_date === 'string') {
        leaveDateStr = leaveDay.leave_date.split('T')[0];
      } else if (leaveDay.leave_date instanceof Date) {
        leaveDateStr = formatDate(leaveDay.leave_date);
      } else {
        return false;
      }
      
      return leaveDateStr === dateStr;
    });

    // Get member names for leave days
    const leaveMemberNames = dayLeaveDays.map((leaveDay) => {
      if (leaveDay.member_name) {
        return leaveDay.member_name;
      }
      // Fallback: find member by ID
      const member = members.find(m => m.id === leaveDay.member_id);
      return member ? member.name : `Member ${leaveDay.member_id}`;
    });

    let classes = 'day';
    if (dayOfWeek === 0 || dayOfWeek === 6) {
      classes += ' weekend';
    }
    if (isHoliday) {
      classes += ' holiday';
    }

    const currentShift = dayShifts.length > 0 ? dayShifts[0] : null;
    const isWorkingDay = dayOfWeek !== 0 && dayOfWeek !== 6 && !isHoliday;

    return (
      <div 
        key={day} 
        className={`${classes} ${isWorkingDay ? 'clickable-day' : ''}`}
        onClick={isWorkingDay ? () => {
          setShowShiftModal(dateStr);
          setSelectedMemberId(currentShift ? currentShift.member_id : '');
          setShiftError('');
        } : undefined}
        title={isWorkingDay ? 'Click to change shift assignment' : ''}
      >
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
        {leaveMemberNames.map((memberName, index) => (
          <span
            key={`leave-${index}`}
            className="leave-name"
            title={`${memberName} is on leave`}
          >
            🏖️ {memberName}
          </span>
        ))}
        {isWorkingDay && dayShifts.length === 0 && (
          <span className="no-shift-indicator" title="No shift assigned - click to assign">
            +
          </span>
        )}
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

      {/* Shift Change Modal */}
      {showShiftModal && (
        <div className="modal-overlay" onClick={() => setShowShiftModal(null)}>
          <div className="modal-content shift-modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <div>
                <h3>Change Shift Assignment</h3>
                <p style={{ margin: '4px 0 0 0', fontSize: '14px', color: '#7f8c8d', fontWeight: 'normal' }}>
                  {new Date(showShiftModal).toLocaleDateString('en-US', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}
                </p>
              </div>
              <button className="modal-close" onClick={() => setShowShiftModal(null)} title="Close">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                  <line x1="18" y1="6" x2="6" y2="18"></line>
                  <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg>
              </button>
            </div>
            
            <div className="modal-body">
              <form onSubmit={async (e) => {
                e.preventDefault();
                setShiftError('');

                if (!selectedMemberId) {
                  setShiftError('Please select a member');
                  return;
                }

                try {
                  await updateShiftForDate(showShiftModal, parseInt(selectedMemberId));
                  setShowShiftModal(null);
                  setSelectedMemberId('');
                } catch (err) {
                  setShiftError(err.message || 'Error updating shift');
                }
              }} className="shift-change-form">
                <div className="form-group">
                  <label htmlFor="shift-member">Select Member</label>
                  <select
                    id="shift-member"
                    value={selectedMemberId}
                    onChange={(e) => setSelectedMemberId(e.target.value)}
                    disabled={loading}
                  >
                    <option value="">-- Select Member --</option>
                    {members
                      .filter(member => {
                        // Exclude members on leave for this date
                        return !leaveDays.some(ld => {
                          let leaveDateStr;
                          if (typeof ld.leave_date === 'string') {
                            leaveDateStr = ld.leave_date.split('T')[0];
                          } else if (ld.leave_date instanceof Date) {
                            leaveDateStr = formatDate(ld.leave_date);
                          } else {
                            return false;
                          }
                          return leaveDateStr === showShiftModal && ld.member_id === member.id;
                        });
                      })
                      .map((member) => (
                        <option key={member.id} value={member.id}>
                          {member.name}
                        </option>
                      ))}
                  </select>
                </div>
                {shiftError && (
                  <div className="error-message" style={{ marginTop: '16px', padding: '12px', background: '#fee', border: '1px solid #fcc', borderRadius: '6px', color: '#c33' }}>
                    {shiftError}
                  </div>
                )}
                <div className="form-actions">
                  <button type="button" className="cancel-btn" onClick={() => setShowShiftModal(null)} disabled={loading}>
                    Cancel
                  </button>
                  <button type="submit" className="save-btn" disabled={loading || !selectedMemberId}>
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                      <polyline points="20 6 9 17 4 12"></polyline>
                    </svg>
                    Save Changes
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}
    </section>
  );
}

