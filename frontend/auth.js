const API_BASE = '/api';

// Get token from localStorage
function getToken() {
    return localStorage.getItem('token');
}

// Save token to localStorage
function setToken(token) {
    localStorage.setItem('token', token);
}

// Remove token
function removeToken() {
    localStorage.removeItem('token');
}

// Tab switching
function showLogin() {
    document.getElementById('loginForm').style.display = 'block';
    document.getElementById('registerForm').style.display = 'none';
    document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
    event.target.classList.add('active');
}

function showRegister() {
    document.getElementById('loginForm').style.display = 'none';
    document.getElementById('registerForm').style.display = 'block';
    document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
    event.target.classList.add('active');
}

// Login
async function login() {
    const username = document.getElementById('loginUsername').value.trim();
    const password = document.getElementById('loginPassword').value;
    const errorDiv = document.getElementById('loginError');

    if (!username || !password) {
        errorDiv.textContent = 'Please enter username and password';
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });

        if (!response.ok) {
            const error = await response.text();
            errorDiv.textContent = error || 'Login failed';
            return;
        }

        const data = await response.json();
        setToken(data.token);
        window.location.href = '/index.html';
    } catch (error) {
        errorDiv.textContent = 'Error logging in: ' + error.message;
    }
}

// Register
async function register() {
    const username = document.getElementById('registerUsername').value.trim();
    const password = document.getElementById('registerPassword').value;
    const passwordConfirm = document.getElementById('registerPasswordConfirm').value;
    const errorDiv = document.getElementById('registerError');

    if (!username || !password) {
        errorDiv.textContent = 'Please enter username and password';
        return;
    }

    if (password !== passwordConfirm) {
        errorDiv.textContent = 'Passwords do not match';
        return;
    }

    if (username.length < 3) {
        errorDiv.textContent = 'Username must be at least 3 characters';
        return;
    }

    if (password.length < 4) {
        errorDiv.textContent = 'Password must be at least 4 characters';
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/auth/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });

        if (!response.ok) {
            const error = await response.text();
            errorDiv.textContent = error || 'Registration failed';
            return;
        }

        const data = await response.json();
        setToken(data.token);
        window.location.href = '/index.html';
    } catch (error) {
        errorDiv.textContent = 'Error registering: ' + error.message;
    }
}
