// Main Application - Reactive UI with State Management

// Check if required objects are available
if (typeof state === 'undefined') {
    console.error('state is not defined. Make sure state.js is loaded.');
}
if (typeof api === 'undefined') {
    console.error('api is not defined. Make sure api.js is loaded.');
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', async () => {
    // Check if state and api are available
    if (typeof state === 'undefined' || typeof api === 'undefined') {
        alert('Error: Required scripts not loaded. Please refresh the page.');
        return;
    }
    
    // Check authentication
    if (!state.get('token')) {
        window.location.href = '/login.html';
        return;
    }

    // Set default dates
    const today = new Date().toISOString().split('T')[0];
    document.getElementById('startDate').value = today;
    const nextMonth = new Date();
    nextMonth.setMonth(nextMonth.getMonth() + 1);
    document.getElementById('endDate').value = nextMonth.toISOString().split('T')[0];

    // Subscribe to state changes FIRST - so updates trigger renders
    setupReactiveUpdates();
    
    // Load initial data - this will trigger subscriptions and render UI
    await loadAllData();
    
    // Setup auto-refresh every 30 seconds
    setupAutoRefresh();
});

// Load all data
async function loadAllData() {
    try {
        // Load holidays first, then others
        await api.getHolidays();
        
        const [members, stats] = await Promise.all([
            api.getMembers().catch(() => []),
            api.getStats().catch(() => [])
        ]);
        
        // Load shifts separately
        await loadShiftsForCurrentMonth().catch(() => {});
        
        // Ensure arrays are not null
        if (!members || !Array.isArray(members)) {
            state.set('members', []);
        }
        if (!stats || !Array.isArray(stats)) {
            state.set('stats', []);
        }
        
        // Force calendar update after all data is loaded
        // Members and stats will be rendered automatically via subscriptions
        updateCalendar();
    } catch (error) {
        console.error('Error in loadAllData:', error);
        if (error.status !== 401) {
            showMessage('Error loading data: ' + (error.message || 'Unknown error'), 'error');
        }
    }
}

// Load shifts for current month
async function loadShiftsForCurrentMonth() {
    const currentMonth = state.get('currentMonth');
    const currentYear = state.get('currentYear');
    const start = new Date(currentYear, currentMonth, 1);
    const end = new Date(currentYear, currentMonth + 1, 0);
    
    try {
        await api.getShifts(formatDate(start), formatDate(end));
    } catch (error) {
        // Set empty array to prevent null errors
        state.set('shifts', []);
    }
}

// Setup reactive UI updates
function setupReactiveUpdates() {
    // Members list updates automatically
    state.subscribe('members', (members) => {
        renderMembers(members);
    });

    // Stats updates automatically
    state.subscribe('stats', (stats) => {
        renderStats(stats);
    });
    
    // Initial render if data already exists
    const existingMembers = state.get('members');
    if (existingMembers && existingMembers.length > 0) {
        renderMembers(existingMembers);
    }
    
    const existingStats = state.get('stats');
    if (existingStats && existingStats.length > 0) {
        renderStats(existingStats);
    }

    // Calendar updates when shifts or month changes
    state.subscribe('shifts', () => {
        updateCalendar();
    });
    
    state.subscribe('holidays', () => {
        updateCalendar();
    });
    
    state.subscribe('currentMonth', () => {
        loadShiftsForCurrentMonth();
    });
    
    state.subscribe('currentYear', () => {
        loadShiftsForCurrentMonth();
    });
}

// Render members list
function renderMembers(members) {
    if (!members || !Array.isArray(members)) {
        return;
    }
    
    const list = document.getElementById('membersList');
    if (!list) return;
    
    if (members.length === 0) {
        list.innerHTML = '<div class="empty-state">No members added yet</div>';
        return;
    }
    
    list.innerHTML = members.map(m => `
        <div class="member-item">
            <span>${escapeHtml(m.name)}</span>
            <button class="delete-btn" onclick="deleteMember(${m.id})">Delete</button>
        </div>
    `).join('');
}

// Render statistics
function renderStats(stats) {
    if (!stats || !Array.isArray(stats)) {
        return;
    }
    
    const list = document.getElementById('statsList');
    if (!list) return;
    
    if (stats.length === 0) {
        list.innerHTML = '<div class="empty-state">No statistics yet</div>';
        return;
    }
    
    list.innerHTML = stats.map(s => `
        <div class="stat-card">
            <h3>${escapeHtml(s.member_name)}</h3>
            <div class="stat-item">
                <span>Total Shift Days:</span>
                <strong>${s.total_days}</strong>
            </div>
            <div class="stat-item">
                <span>Long Shift Count:</span>
                <strong>${s.long_shift_count}</strong>
            </div>
        </div>
    `).join('');
}

// Add member
async function addMember() {
    if (typeof api === 'undefined') {
        alert('Error: API not loaded. Please refresh the page.');
        return;
    }
    
    const nameInput = document.getElementById('memberName');
    if (!nameInput) {
        alert('Error: Member name input not found.');
        return;
    }
    
    const name = nameInput.value.trim();
    
    if (!name) {
        showMessage('Please enter member name', 'error');
        return;
    }
    
    try {
        await api.createMember(name);
        nameInput.value = '';
        showMessage('Member added successfully', 'success');
        // Stats will update automatically via subscription
        await api.getStats();
    } catch (error) {
        showMessage('Error adding member: ' + (error.message || 'Unknown error'), 'error');
    }
}

// Delete member
async function deleteMember(id) {
    if (typeof api === 'undefined') {
        alert('Error: API not loaded. Please refresh the page.');
        return;
    }
    
    if (!confirm('Are you sure you want to delete this member?')) {
        return;
    }
    
    try {
        await api.deleteMember(id);
        showMessage('Member deleted successfully', 'success');
        // Stats and shifts will update automatically
        await api.getStats();
        await loadShiftsForCurrentMonth();
    } catch (error) {
        showMessage('Error deleting member: ' + (error.message || 'Unknown error'), 'error');
    }
}

// Generate shift plan
async function generateShifts() {
    if (typeof api === 'undefined') {
        alert('Error: API not loaded. Please refresh the page.');
        return;
    }
    
    const startDateInput = document.getElementById('startDate');
    const endDateInput = document.getElementById('endDate');
    
    if (!startDateInput || !endDateInput) {
        alert('Error: Date inputs not found.');
        return;
    }
    
    const startDate = startDateInput.value;
    const endDate = endDateInput.value;
    
    if (!startDate || !endDate) {
        showMessage('Please select start and end dates', 'error');
        return;
    }
    
    if (new Date(startDate) > new Date(endDate)) {
        showMessage('Start date cannot be after end date', 'error');
        return;
    }
    
    try {
        const shifts = await api.generateShifts(startDate, endDate);
        
        // Reload shifts for current month to update calendar
        await loadShiftsForCurrentMonth();
        
        // Also reload stats
        await api.getStats();
        
        showMessage(`${shifts.length} shift plans created`, 'success');
        // Calendar will update automatically via subscription
    } catch (error) {
        showMessage('Error creating plan: ' + (error.message || 'Unknown error'), 'error');
    }
}

// Update calendar
function updateCalendar() {
    if (typeof state === 'undefined') {
        return;
    }
    
    const calendar = document.getElementById('calendar');
    if (!calendar) {
        return;
    }
    
    const currentMonth = state.get('currentMonth');
    const currentYear = state.get('currentYear');
    const shifts = state.get('shifts') || [];
    const holidays = state.get('holidays') || {};
    
    
    const firstDay = new Date(currentYear, currentMonth, 1);
    const lastDay = new Date(currentYear, currentMonth + 1, 0);
    const daysInMonth = lastDay.getDate();
    
    // getDay() returns 0=Sunday, 1=Monday, etc. We want Monday=0
    let startingDayOfWeek = firstDay.getDay() - 1;
    if (startingDayOfWeek < 0) startingDayOfWeek = 6; // Sunday becomes 6
    
    const monthNames = [
        'January', 'February', 'March', 'April', 'May', 'June',
        'July', 'August', 'September', 'October', 'November', 'December'
    ];
    
    document.getElementById('currentMonth').textContent = 
        `${monthNames[currentMonth]} ${currentYear}`;

    const weekDays = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
    
    let html = '';
    
    // Week day headers
    weekDays.forEach(day => {
        html += `<div class="calendar-header">${day}</div>`;
    });
    
    // Empty days (before first day of month)
    for (let i = 0; i < startingDayOfWeek; i++) {
        html += '<div class="day empty"></div>';
    }
    
    // Days of the month
    for (let day = 1; day <= daysInMonth; day++) {
        const date = new Date(currentYear, currentMonth, day);
        const dateStr = formatDate(date);
        
        // Filter shifts for this day
        const dayShifts = (shifts || []).filter(s => {
            if (!s.start_date || !s.end_date) {
                // Skip shifts with invalid dates
                return false;
            }
            
            // Handle different date formats
            let startDateStr, endDateStr;
            
            if (typeof s.start_date === 'string') {
                startDateStr = s.start_date.split('T')[0];
            } else if (s.start_date instanceof Date) {
                startDateStr = formatDate(s.start_date);
            } else {
                // Try to parse as ISO string
                startDateStr = String(s.start_date).split('T')[0];
            }
            
            if (typeof s.end_date === 'string') {
                endDateStr = s.end_date.split('T')[0];
            } else if (s.end_date instanceof Date) {
                endDateStr = formatDate(s.end_date);
            } else {
                // Try to parse as ISO string
                endDateStr = String(s.end_date).split('T')[0];
            }
            
            // Skip if dates are invalid (zero dates)
            if (startDateStr === '0001-01-01' || endDateStr === '0001-01-01') {
                return false;
            }
            
            return dateStr >= startDateStr && dateStr <= endDateStr;
        });
        
        let classes = 'day';
        const dayOfWeek = date.getDay();
        if (dayOfWeek === 0 || dayOfWeek === 6) {
            classes += ' weekend';
        }
        
        // Check if it's a holiday - try multiple date formats
        let holidayName = holidays[dateStr];
        if (!holidayName) {
            // Try alternative formats
            const altDateStr = date.toISOString().split('T')[0];
            holidayName = holidays[altDateStr];
        }
        const isHoliday = !!holidayName;
        
        if (isHoliday) {
            classes += ' holiday';
        }
        
        html += `<div class="${classes}">`;
        const dayNumberClass = isHoliday ? 'day-number holiday-day' : 'day-number';
        html += `<span class="${dayNumberClass}">${day}</span>`;
        
        // Show holiday name if it's a holiday
        if (isHoliday) {
            html += `<span class="holiday-name" title="${escapeHtml(holidayName)}">${escapeHtml(holidayName)}</span>`;
        }
        
        // Show shifts at the bottom
        if (dayShifts.length > 0) {
            dayShifts.forEach(shift => {
                const shiftClass = shift.is_long_shift ? 'long-shift' : '';
                const memberName = shift.member_name || 'Unknown';
                html += `<span class="shift-name ${shiftClass}">${escapeHtml(memberName)}</span>`;
            });
        }
        
        html += '</div>';
    }
    
    calendar.innerHTML = html;
}

// Previous month
function previousMonth() {
    if (typeof state === 'undefined') {
        alert('Error: State not loaded. Please refresh the page.');
        return;
    }
    
    const currentMonth = state.get('currentMonth');
    const currentYear = state.get('currentYear');
    
    let newMonth = currentMonth - 1;
    let newYear = currentYear;
    
    if (newMonth < 0) {
        newMonth = 11;
        newYear--;
    }
    
    state.update({
        currentMonth: newMonth,
        currentYear: newYear
    });
}

// Next month
function nextMonth() {
    if (typeof state === 'undefined') {
        alert('Error: State not loaded. Please refresh the page.');
        return;
    }
    
    const currentMonth = state.get('currentMonth');
    const currentYear = state.get('currentYear');
    
    let newMonth = currentMonth + 1;
    let newYear = currentYear;
    
    if (newMonth > 11) {
        newMonth = 0;
        newYear++;
    }
    
    state.update({
        currentMonth: newMonth,
        currentYear: newYear
    });
}

// Logout
async function logout() {
    if (typeof api === 'undefined') {
        // Even if API is not loaded, redirect to login
        window.location.href = '/login.html';
        return;
    }
    
    try {
        await api.logout();
    } catch (error) {
        // Ignore errors on logout
    } finally {
        window.location.href = '/login.html';
    }
}

// Helper functions
function formatDate(date) {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Setup auto-refresh for data
function setupAutoRefresh() {
    // Refresh data every 30 seconds
    setInterval(async () => {
        try {
            await Promise.all([
                api.getMembers(),
                api.getStats(),
                loadShiftsForCurrentMonth()
            ]);
        } catch (error) {
            // Silently fail - session might have expired
            if (error.status === 401) {
                // Session expired, will be handled by API wrapper
                return;
            }
        }
    }, 30000); // 30 seconds
}

function showMessage(message, type) {
    const existing = document.querySelector('.message');
    if (existing) {
        existing.remove();
    }
    
    const msg = document.createElement('div');
    msg.className = `message ${type}`;
    msg.textContent = message;
    
    const container = document.querySelector('.main-content');
    if (container) {
        container.insertBefore(msg, container.firstChild);
        setTimeout(() => msg.remove(), 5000);
    }
}
