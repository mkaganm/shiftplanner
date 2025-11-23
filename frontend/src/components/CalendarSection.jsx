// Calendar Section Component
import { useState } from 'react';
import { useApp } from '../context/AppContext';
import './CalendarSection.css';

export function CalendarSection() {
  const { currentMonth, currentYear, shifts, holidays, leaveDays, members, previousMonth, nextMonth, formatDate, updateShiftForDate, loading } = useApp();
  const [showShiftModal, setShowShiftModal] = useState(null);
  const [selectedMemberId, setSelectedMemberId] = useState('');
  const [shiftError, setShiftError] = useState('');
  const [showExportMenu, setShowExportMenu] = useState(false);

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
      
      // Check if member still exists
      const memberExists = members.some(m => m.id === leaveDay.member_id);
      if (!memberExists) {
        return false; // Skip leave days for deleted members
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

  const prepareData = () => {
    // Prepare data rows
    const rows = [];
    
    // Header row
    rows.push(['Date', 'Day', 'Member Name', 'Shift Type', 'Is Holiday']);
    
    // Get all days in the current month
    const daysInMonth = new Date(currentYear, currentMonth + 1, 0).getDate();
    
    for (let day = 1; day <= daysInMonth; day++) {
      const date = new Date(currentYear, currentMonth, day);
      const dateStr = formatDate(date);
      const dayOfWeek = date.getDay();
      const dayNames = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
      
      // Check if it's a holiday
      const holidayName = holidays[dateStr];
      const isHoliday = !!holidayName;
      
      // Filter shifts for this day
      const dayShifts = shifts.filter((shift) => {
        if (!shift || !shift.start_date || !shift.end_date) {
          return false;
        }
        
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
        
        // Check if member still exists
        const memberExists = members.some(m => m.id === leaveDay.member_id);
        if (!memberExists) {
          return false; // Skip leave days for deleted members
        }
        
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
        const member = members.find(m => m.id === leaveDay.member_id);
        return member ? member.name : `Member ${leaveDay.member_id}`;
      });
      
      if (dayShifts.length > 0) {
        // Add row for each shift
        dayShifts.forEach((shift) => {
          rows.push([
            dateStr,
            dayNames[dayOfWeek],
            shift.member_name || 'Unknown',
            shift.is_long_shift ? 'Long Shift' : 'Normal Shift',
            isHoliday ? holidayName : ''
          ]);
        });
      } else if (leaveMemberNames.length > 0) {
        // Add row for leave days
        leaveMemberNames.forEach((memberName) => {
          rows.push([
            dateStr,
            dayNames[dayOfWeek],
            memberName,
            'On Leave',
            isHoliday ? holidayName : ''
          ]);
        });
      } else if (isHoliday) {
        // Add row for holidays without shifts
        rows.push([
          dateStr,
          dayNames[dayOfWeek],
          '',
          'Holiday',
          holidayName
        ]);
      } else if (dayOfWeek !== 0 && dayOfWeek !== 6) {
        // Add row for working days without shifts
        rows.push([
          dateStr,
          dayNames[dayOfWeek],
          '',
          'No Shift',
          ''
        ]);
      }
    }
    
    return rows;
  };

  const exportToCSV = () => {
    const rows = prepareData();
    
    // Convert to CSV string
    const csvContent = rows.map(row => {
      return row.map(cell => {
        // Escape quotes and wrap in quotes if contains comma, quote, or newline
        if (typeof cell === 'string' && (cell.includes(',') || cell.includes('"') || cell.includes('\n'))) {
          return `"${cell.replace(/"/g, '""')}"`;
        }
        return cell || '';
      }).join(',');
    }).join('\n');
    
    // Create blob and download
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    const url = URL.createObjectURL(blob);
    
    link.setAttribute('href', url);
    link.setAttribute('download', `shift-plan-${monthNames[currentMonth]}-${currentYear}.csv`);
    link.style.visibility = 'hidden';
    
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    setShowExportMenu(false);
  };

  const exportToJSON = () => {
    const rows = prepareData();
    const headers = rows[0];
    const data = rows.slice(1).map(row => {
      const obj = {};
      headers.forEach((header, index) => {
        obj[header] = row[index];
      });
      return obj;
    });
    
    const jsonContent = JSON.stringify({
      month: monthNames[currentMonth],
      year: currentYear,
      data: data
    }, null, 2);
    
    const blob = new Blob([jsonContent], { type: 'application/json;charset=utf-8;' });
    const link = document.createElement('a');
    const url = URL.createObjectURL(blob);
    
    link.setAttribute('href', url);
    link.setAttribute('download', `shift-plan-${monthNames[currentMonth]}-${currentYear}.json`);
    link.style.visibility = 'hidden';
    
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    setShowExportMenu(false);
  };

  const exportToExcel = () => {
    const rows = prepareData();
    
    // Create HTML table for Excel
    let html = '<html><head><meta charset="utf-8"><style>table{border-collapse:collapse;}th,td{border:1px solid #ddd;padding:8px;text-align:left;}th{background-color:#34495e;color:white;font-weight:bold;}</style></head><body><table>';
    
    rows.forEach((row, index) => {
      html += '<tr>';
      row.forEach(cell => {
        const tag = index === 0 ? 'th' : 'td';
        html += `<${tag}>${cell || ''}</${tag}>`;
      });
      html += '</tr>';
    });
    
    html += '</table></body></html>';
    
    const blob = new Blob([html], { type: 'application/vnd.ms-excel' });
    const link = document.createElement('a');
    const url = URL.createObjectURL(blob);
    
    link.setAttribute('href', url);
    link.setAttribute('download', `shift-plan-${monthNames[currentMonth]}-${currentYear}.xls`);
    link.style.visibility = 'hidden';
    
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    setShowExportMenu(false);
  };

  return (
    <section className="section">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
        <h2 style={{ margin: 0 }}>Shift Plan</h2>
        <div style={{ position: 'relative' }}>
          <button onClick={() => setShowExportMenu(!showExportMenu)} className="export-btn" title="Export">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
              <polyline points="7 10 12 15 17 10"></polyline>
              <line x1="12" y1="15" x2="12" y2="3"></line>
            </svg>
            Export
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" style={{ marginLeft: '4px' }}>
              <polyline points="6 9 12 15 18 9"></polyline>
            </svg>
          </button>
          {showExportMenu && (
            <div className="export-menu">
              <button onClick={exportToCSV} className="export-menu-item">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                  <polyline points="7 10 12 15 17 10"></polyline>
                  <line x1="12" y1="15" x2="12" y2="3"></line>
                </svg>
                Export as CSV
              </button>
              <button onClick={exportToExcel} className="export-menu-item">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
                  <line x1="9" y1="3" x2="9" y2="21"></line>
                  <line x1="3" y1="9" x2="21" y2="9"></line>
                </svg>
                Export as Excel
              </button>
              <button onClick={exportToJSON} className="export-menu-item">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"></path>
                </svg>
                Export as JSON
              </button>
            </div>
          )}
        </div>
      </div>
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

      {/* Click outside to close export menu */}
      {showExportMenu && (
        <div 
          style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, zIndex: 999 }}
          onClick={() => setShowExportMenu(false)}
        />
      )}

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

