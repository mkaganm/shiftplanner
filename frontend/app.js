const API_BASE = '/api';

let currentMonth = new Date().getMonth();
let currentYear = new Date().getFullYear();

// Token check
function getToken() {
    return localStorage.getItem('token');
}

// Add headers for API requests
function getHeaders() {
    const headers = { 'Content-Type': 'application/json' };
    const token = getToken();
    if (token) {
        headers['Authorization'] = token;
    }
    return headers;
}

// On page load
document.addEventListener('DOMContentLoaded', () => {
    // Token check
    if (!getToken()) {
        window.location.href = '/login.html';
        return;
    }

    loadMembers();
    loadStats();
    // Load holidays first, then shifts (which will also load holidays and update calendar)
    loadHolidays().then(() => {
        loadShifts();
    });
    
    // Set today's date as default
    const today = new Date().toISOString().split('T')[0];
    document.getElementById('startDate').value = today;
    const nextMonth = new Date();
    nextMonth.setMonth(nextMonth.getMonth() + 1);
    document.getElementById('endDate').value = nextMonth.toISOString().split('T')[0];
});

// Logout
async function logout() {
    const token = getToken();
    if (token) {
        try {
            await fetch(`${API_BASE}/auth/logout`, {
                method: 'POST',
                headers: getHeaders()
            });
        } catch (error) {
            console.error('Logout error:', error);
        }
    }
    localStorage.removeItem('token');
    window.location.href = '/login.html';
}

// Add member
async function addMember() {
    const nameInput = document.getElementById('memberName');
    const name = nameInput.value.trim();
    
    if (!name) {
        showMessage('Please enter member name', 'error');
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE}/members`, {
            method: 'POST',
            headers: getHeaders(),
            body: JSON.stringify({ name })
        });
        
        if (!response.ok) throw new Error('Failed to add member');
        
        nameInput.value = '';
        loadMembers();
        loadStats();
        showMessage('Member added successfully', 'success');
    } catch (error) {
        showMessage('Error adding member: ' + error.message, 'error');
    }
}

// Delete member
async function deleteMember(id) {
    if (!confirm('Are you sure you want to delete this member?')) {
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE}/members/${id}`, {
            method: 'DELETE',
            headers: getHeaders()
        });
        
        if (!response.ok) throw new Error('Failed to delete member');
        
        loadMembers();
        loadStats();
        loadShifts();
        showMessage('Member deleted successfully', 'success');
    } catch (error) {
        showMessage('Error deleting member: ' + error.message, 'error');
    }
}

// Load members
async function loadMembers() {
    try {
        const response = await fetch(`${API_BASE}/members`, {
            headers: getHeaders()
        });
        
        if (response.status === 401) {
            localStorage.removeItem('token');
            window.location.href = '/login.html';
            return;
        }
        
        if (!response.ok) {
            throw new Error('Failed to load members');
        }
        
        const members = await response.json();
        
        const list = document.getElementById('membersList');
        if (members.length === 0) {
            list.innerHTML = '<div class="empty-state">No members added yet</div>';
            return;
        }
        
        list.innerHTML = members.map(m => `
            <div class="member-card">
                <span>${escapeHtml(m.name)}</span>
                <button class="delete-btn" onclick="deleteMember(${m.id})">Delete</button>
            </div>
        `).join('');
    } catch (error) {
        showMessage('Error loading members: ' + error.message, 'error');
    }
}

// Load statistics
async function loadStats() {
    try {
        const response = await fetch(`${API_BASE}/stats`, {
            headers: getHeaders()
        });
        
        if (response.status === 401) {
            return;
        }
        
        const stats = await response.json();
        
        const list = document.getElementById('statsList');
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
    } catch (error) {
        console.error('Error loading statistics:', error);
    }
}

// Generate shift plan
async function generateShifts() {
    const startDate = document.getElementById('startDate').value;
    const endDate = document.getElementById('endDate').value;
    
    if (!startDate || !endDate) {
        showMessage('Please select start and end dates', 'error');
        return;
    }
    
    if (new Date(startDate) > new Date(endDate)) {
        showMessage('Start date cannot be after end date', 'error');
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE}/shifts/generate`, {
            method: 'POST',
            headers: getHeaders(),
            body: JSON.stringify({ start_date: startDate, end_date: endDate })
        });
        
        if (response.status === 401) {
            localStorage.removeItem('token');
            window.location.href = '/login.html';
            return;
        }
        
        if (!response.ok) {
            const errorText = await response.text();
            let errorMsg = 'Failed to create plan';
            try {
                const errorJson = JSON.parse(errorText);
                errorMsg = errorJson.error || errorMsg;
            } catch {
                errorMsg = errorText || errorMsg;
            }
            throw new Error(errorMsg);
        }
        
        const shifts = await response.json();
        // Reload shifts to update calendar
        await loadShifts();
        loadStats();
        showMessage(`${shifts.length} shift plans created`, 'success');
    } catch (error) {
        showMessage('Error creating plan: ' + error.message, 'error');
    }
}

// Load shifts
let shifts = [];
let holidays = {};

async function loadShifts() {
    try {
        const start = new Date(currentYear, currentMonth, 1);
        const end = new Date(currentYear, currentMonth + 1, 0);
        
        const response = await fetch(
            `${API_BASE}/shifts?start_date=${formatDate(start)}&end_date=${formatDate(end)}`,
            { headers: getHeaders() }
        );
        
        if (response.status === 401) {
            return;
        }
        
        shifts = await response.json();
        
        // Ensure holidays are loaded
        if (Object.keys(holidays).length === 0) {
            await loadHolidays();
        }
        
        updateCalendar();
    } catch (error) {
        console.error('Error loading shifts:', error);
    }
}

// Load holidays from API
async function loadHolidays() {
    try {
        const response = await fetch(`${API_BASE}/holidays`);
        if (response.ok) {
            holidays = await response.json();
        }
    } catch (error) {
        console.error('Error loading holidays:', error);
    }
}

// Update calendar
function updateCalendar() {
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
    
    const calendar = document.getElementById('calendar');
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
        
        // Filter shifts for this day - compare date strings directly
        const dayShifts = shifts.filter(s => {
            if (!s.start_date || !s.end_date) return false;
            
            // Extract date part from ISO string (e.g., "2025-01-01T00:00:00Z" -> "2025-01-01")
            const startDateStr = String(s.start_date).split('T')[0];
            const endDateStr = String(s.end_date).split('T')[0];
            
            // Check if current date is within shift range
            return dateStr >= startDateStr && dateStr <= endDateStr;
        });
        
        let classes = 'day';
        const dayOfWeek = date.getDay();
        if (dayOfWeek === 0 || dayOfWeek === 6) {
            classes += ' weekend';
        }
        
        // Check if it's a holiday from API
        const holidayName = holidays[dateStr];
        const isHoliday = !!holidayName;
        if (isHoliday) {
            classes += ' holiday';
        }
        
        html += `<div class="${classes}">`;
        html += `<span class="day-number">${day}</span>`;
        
        // Show holiday name if it's a holiday
        if (isHoliday) {
            html += `<span class="holiday-name">${escapeHtml(holidayName)}</span>`;
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
    currentMonth--;
    if (currentMonth < 0) {
        currentMonth = 11;
        currentYear--;
    }
    loadShifts();
}

// Next month
function nextMonth() {
    currentMonth++;
    if (currentMonth > 11) {
        currentMonth = 0;
        currentYear++;
    }
    loadShifts();
}

// Helper functions
function formatDate(date) {
    // Format date as YYYY-MM-DD in local timezone
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

function showMessage(message, type) {
    const existing = document.querySelector('.message');
    if (existing) {
        existing.remove();
    }
    
    const msg = document.createElement('div');
    msg.className = `message ${type}`;
    msg.textContent = message;
    
    const container = document.querySelector('.main-content');
    container.insertBefore(msg, container.firstChild);
    
    setTimeout(() => msg.remove(), 5000);
}

