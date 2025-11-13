// Authentication - Using API wrapper

// Tab switching
function showLogin(event) {
    document.getElementById('loginForm').style.display = 'block';
    document.getElementById('registerForm').style.display = 'none';
    document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
    if (event && event.target) {
        event.target.classList.add('active');
    } else {
        document.querySelectorAll('.tab-btn')[0].classList.add('active');
    }
}

function showRegister(event) {
    document.getElementById('loginForm').style.display = 'none';
    document.getElementById('registerForm').style.display = 'block';
    document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
    if (event && event.target) {
        event.target.classList.add('active');
    } else {
        document.querySelectorAll('.tab-btn')[1].classList.add('active');
    }
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

    // Check if api object is available
    if (typeof api === 'undefined') {
        errorDiv.textContent = 'Error: API not loaded. Please refresh the page.';
        console.error('api object is not defined');
        return;
    }

    try {
        errorDiv.textContent = 'Logging in...';
        await api.login(username, password);
        window.location.href = '/index.html';
    } catch (error) {
        console.error('Login error:', error);
        errorDiv.textContent = error.message || 'Login failed';
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

    // Check if api object is available
    if (typeof api === 'undefined') {
        errorDiv.textContent = 'Error: API not loaded. Please refresh the page.';
        console.error('api object is not defined');
        return;
    }

    try {
        errorDiv.textContent = 'Registering...';
        await api.register(username, password);
        window.location.href = '/index.html';
    } catch (error) {
        console.error('Registration error:', error);
        errorDiv.textContent = error.message || 'Registration failed';
    }
}
