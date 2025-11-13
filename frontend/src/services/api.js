// API Service Layer
const API_BASE = '/api';

class APIError extends Error {
  constructor(message, status) {
    super(message);
    this.status = status;
    this.name = 'APIError';
  }
}

// Helper function for API requests
async function apiRequest(endpoint, options = {}) {
  const token = localStorage.getItem('token');
  
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
    
    // Handle 401 Unauthorized
    if (response.status === 401) {
      localStorage.removeItem('token');
      throw new APIError('Session expired. Please login again.', 401);
    }
    
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
    
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      return await response.json();
    }
    
    return null;
  } catch (error) {
    if (error instanceof APIError) {
      throw error;
    }
    throw new APIError(error.message || 'Network error', 0);
  }
}

// Auth API
export const authAPI = {
  async register(username, password) {
    const data = await apiRequest('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ username, password })
    });
    if (data.token) {
      localStorage.setItem('token', data.token);
    }
    return data;
  },
  
  async login(username, password) {
    const data = await apiRequest('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password })
    });
    if (data.token) {
      localStorage.setItem('token', data.token);
    }
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
      localStorage.removeItem('token');
    }
  }
};

// Members API
export const membersAPI = {
  async getAll() {
    return await apiRequest('/members');
  },
  
  async create(name) {
    return await apiRequest('/members', {
      method: 'POST',
      body: JSON.stringify({ name })
    });
  },
  
  async delete(id) {
    return await apiRequest(`/members/${id}`, {
      method: 'DELETE'
    });
  }
};

// Shifts API
export const shiftsAPI = {
  async get(startDate, endDate) {
    const params = new URLSearchParams({
      start_date: startDate,
      end_date: endDate
    });
    return await apiRequest(`/shifts?${params}`);
  },
  
  async generate(startDate, endDate) {
    return await apiRequest('/shifts/generate', {
      method: 'POST',
      body: JSON.stringify({ start_date: startDate, end_date: endDate })
    });
  }
};

// Holidays API
export const holidaysAPI = {
  async getAll() {
    return await apiRequest('/holidays');
  }
};

// Stats API
export const statsAPI = {
  async getAll() {
    return await apiRequest('/stats');
  }
};

