// API Wrapper with automatic session handling
const API_BASE = '/api';

class APIError extends Error {
    constructor(message, status) {
        super(message);
        this.status = status;
        this.name = 'APIError';
    }
}

// API request wrapper with automatic session handling
async function apiRequest(endpoint, options = {}) {
    const token = state.get('token');
    
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers
    };
    
    if (token) {
        headers['Authorization'] = token;
    }
    
    try {
        const response = await fetch(`${API_BASE}${endpoint}`, {
            ...options,
            headers
        });
        
        // Handle 401 Unauthorized - session expired
        if (response.status === 401) {
            state.set('token', null);
            state.set('user', null);
            
            // Only redirect if we're not already on login page
            if (!window.location.pathname.includes('login.html')) {
                window.location.href = '/login.html';
            }
            throw new APIError('Session expired. Please login again.', 401);
        }
        
        // Handle other errors
        if (!response.ok) {
            const errorText = await response.text();
            let errorMsg = `Request failed: ${response.status}`;
            try {
                const errorJson = JSON.parse(errorText);
                errorMsg = errorJson.error || errorMsg;
            } catch {
                errorMsg = errorText || errorMsg;
            }
            throw new APIError(errorMsg, response.status);
        }
        
        // Parse JSON response
        const contentType = response.headers.get('content-type');
        if (contentType && contentType.includes('application/json')) {
            return await response.json();
        }
        
        return await response.text();
    } catch (error) {
        if (error instanceof APIError) {
            throw error;
        }
        throw new APIError(`Network error: ${error.message}`, 0);
    }
}

// API methods
const api = {
    // Auth
    async register(username, password) {
        const data = await apiRequest('/auth/register', {
            method: 'POST',
            body: JSON.stringify({ username, password })
        });
        state.set('token', data.token);
        state.set('user', data.user);
        return data;
    },
    
    async login(username, password) {
        const data = await apiRequest('/auth/login', {
            method: 'POST',
            body: JSON.stringify({ username, password })
        });
        state.set('token', data.token);
        state.set('user', data.user);
        return data;
    },
    
    async logout() {
        try {
            await apiRequest('/auth/logout', {
                method: 'POST'
            });
        } catch (error) {
            // Ignore errors on logout
        } finally {
            state.set('token', null);
            state.set('user', null);
        }
    },
    
    // Members
    async getMembers() {
        const members = await apiRequest('/members');
        const membersArray = Array.isArray(members) ? members : [];
        state.set('members', membersArray);
        return membersArray;
    },
    
    async createMember(name) {
        const member = await apiRequest('/members', {
            method: 'POST',
            body: JSON.stringify({ name })
        });
        // Refresh members list
        await api.getMembers();
        return member;
    },
    
    async deleteMember(id) {
        await apiRequest(`/members/${id}`, {
            method: 'DELETE'
        });
        // Refresh members list
        await api.getMembers();
    },
    
    // Shifts
    async getShifts(startDate, endDate) {
        const params = new URLSearchParams({
            start_date: startDate,
            end_date: endDate
        });
        const shifts = await apiRequest(`/shifts?${params}`);
        const shiftsArray = Array.isArray(shifts) ? shifts : [];
        state.set('shifts', shiftsArray);
        return shiftsArray;
    },
    
    async generateShifts(startDate, endDate) {
        const shifts = await apiRequest('/shifts/generate', {
            method: 'POST',
            body: JSON.stringify({ start_date: startDate, end_date: endDate })
        });
        // Refresh shifts
        const currentMonth = state.get('currentMonth');
        const currentYear = state.get('currentYear');
        const start = new Date(currentYear, currentMonth, 1);
        const end = new Date(currentYear, currentMonth + 1, 0);
        await api.getShifts(formatDate(start), formatDate(end));
        return shifts;
    },
    
    // Holidays
    async getHolidays() {
        const holidays = await apiRequest('/holidays');
        const holidaysObj = holidays && typeof holidays === 'object' ? holidays : {};
        state.set('holidays', holidaysObj);
        return holidaysObj;
    },
    
    // Stats
    async getStats() {
        const stats = await apiRequest('/stats');
        const statsArray = Array.isArray(stats) ? stats : [];
        state.set('stats', statsArray);
        return statsArray;
    }
};

// Helper function for date formatting
function formatDate(date) {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
}

