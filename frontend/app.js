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
    loadShifts();
    updateCalendar();
    
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
        loadShifts();
        loadStats();
        showMessage(`${shifts.length} shift plans created`, 'success');
    } catch (error) {
        showMessage('Error creating plan: ' + error.message, 'error');
    }
}

// Load shifts
let shifts = [];
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
        updateCalendar();
    } catch (error) {
        console.error('Error loading shifts:', error);
    }
}

// Update calendar
function updateCalendar() {
    const firstDay = new Date(currentYear, currentMonth, 1);
    const lastDay = new Date(currentYear, currentMonth + 1, 0);
    const daysInMonth = lastDay.getDate();
    const startingDayOfWeek = firstDay.getDay();
    
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
        html += '<div class="calendar-day"></div>';
    }
    
    // Days of the month
    for (let day = 1; day <= daysInMonth; day++) {
        const date = new Date(currentYear, currentMonth, day);
        const dateStr = formatDate(date);
        
        const dayShifts = shifts.filter(s => {
            const start = new Date(s.start_date);
            const end = new Date(s.end_date);
            return date >= start && date <= end;
        });
        
        let classes = 'calendar-day';
        if (date.getDay() === 0 || date.getDay() === 6) {
            classes += ' weekend';
        }
        
        // Holiday check (simple - should be fetched from API in real app)
        const isHoliday = checkHoliday(date);
        if (isHoliday) {
            classes += ' holiday';
        }
        
        if (dayShifts.length > 0) {
            classes += ' has-shift';
        }
        
        html += `<div class="${classes}">`;
        html += `<div class="day-number">${day}</div>`;
        
        if (dayShifts.length > 0) {
            dayShifts.forEach(shift => {
                const shiftClass = shift.is_long_shift ? 'long-shift' : '';
                html += `<div class="shift-info ${shiftClass}">${escapeHtml(shift.member_name)}${shift.is_long_shift ? ' (Long)' : ''}</div>`;
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
    return date.toISOString().split('T')[0];
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

function checkHoliday(date) {
    // Simple holiday check - should be fetched from API in real app
    const month = date.getMonth();
    const day = date.getDate();
    
    // New Year's Day
    if (month === 0 && day === 1) return true;
    // April 23
    if (month === 3 && day === 23) return true;
    // May 1
    if (month === 4 && day === 1) return true;
    // May 19
    if (month === 4 && day === 19) return true;
    // August 30
    if (month === 7 && day === 30) return true;
    // October 29
    if (month === 9 && day === 29) return true;
    
    return false;
}
