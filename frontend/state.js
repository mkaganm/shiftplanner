// Simple State Management System
class StateManager {
    constructor() {
        this.state = {
            members: [],
            shifts: [],
            holidays: {},
            stats: [],
            currentMonth: new Date().getMonth(),
            currentYear: new Date().getFullYear(),
            user: null,
            token: localStorage.getItem('token') || null
        };
        this.listeners = {};
    }

    // Get state value
    get(key) {
        return this.state[key];
    }

    // Set state value and notify listeners
    set(key, value) {
        const oldValue = this.state[key];
        this.state[key] = value;
        
        // Notify listeners
        if (this.listeners[key]) {
            this.listeners[key].forEach(callback => {
                callback(value, oldValue);
            });
        }
        
        // Notify global listeners
        if (this.listeners['*']) {
            this.listeners['*'].forEach(callback => {
                callback(key, value, oldValue);
            });
        }
    }

    // Subscribe to state changes
    subscribe(key, callback) {
        if (!this.listeners[key]) {
            this.listeners[key] = [];
        }
        this.listeners[key].push(callback);
        
        // Return unsubscribe function
        return () => {
            this.listeners[key] = this.listeners[key].filter(cb => cb !== callback);
        };
    }

    // Update multiple values at once
    update(updates) {
        Object.keys(updates).forEach(key => {
            this.set(key, updates[key]);
        });
    }
}

// Global state instance
const state = new StateManager();

// Auto-save token to localStorage when it changes
state.subscribe('token', (token) => {
    if (token) {
        localStorage.setItem('token', token);
    } else {
        localStorage.removeItem('token');
    }
});

