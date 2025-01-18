// Admin token management
class AdminAuth {
    static TOKEN_KEY = 'adminToken';

    static getToken() {
        return localStorage.getItem(this.TOKEN_KEY);
    }

    static setToken(token) {
        localStorage.setItem(this.TOKEN_KEY, token);
    }

    static removeToken() {
        localStorage.removeItem(this.TOKEN_KEY);
    }

    static isAuthenticated() {
        return !!this.getToken();
    }
}




// API calls with error handling
class AdminAPI {
    static async fetchWithAuth(endpoint, options = {}) {
        const token = AdminAuth.getToken();
        if (!token) {
            throw new Error('No authentication token');
        }

        const response = await fetch(endpoint, {
            ...options,
            headers: {
                ...options.headers,
                'Authorization': token,
                'Content-Type': 'application/json',
            }
        });

        if (!response.ok) {
            if (response.status === 401 || response.status === 403) {
                AdminAuth.removeToken();
                showLoginForm();
                throw new Error('Authentication failed');
            }
            throw new Error(`API call failed: ${response.statusText}`);
        }

        return response.json();
    }

    static async getStats() {
        return this.fetchWithAuth('/api/admin/stats');
    }

    static async broadcastEmailToSelectedUsers(subject, body, users) {
        return this.fetchWithAuth('/api/admin/broadcast/selected', {
            method: 'POST',
            body: JSON.stringify({ subject, body, users })
        });
    }

    static async getUsers() {
        return this.fetchWithAuth('/api/admin/users');
    }

    static async deleteUser(userId) {
        return this.fetchWithAuth(`/api/admin/users/delete?id=${userId}`, {
            method: 'POST',
        });
    }
    static async editUser(userId, username, email) {
        return this.fetchWithAuth(`/api/admin/users/edit`, {
            method: 'POST',
            body: JSON.stringify({ id: userId, username, email })
        });
    }
    static async sendBroadcastEmail(recipient, subject, body) {
        return this.fetchWithAuth('/api/admin/broadcast', {
            method: 'POST',
            body: JSON.stringify({ recipient, subject, body })
        });
    }
}

// UI Management
function showAdminContent() {
    document.getElementById('admin-content').style.display = 'block';
    document.getElementById('login-form').style.display = 'none';
    loadDashboard();
}

function showLoginForm() {
    document.getElementById('admin-content').style.display = 'none';
    document.getElementById('login-form').style.display = 'block';
}

async function loadDashboard() {
    try {
        const stats = await AdminAPI.getStats();
        updateStatsDisplay(stats);
        await loadUsersList();
    } catch (error) {
        console.error('Error loading dashboard:', error);
        showError('Failed to load dashboard data');
    }
}

function updateStatsDisplay(stats) {
    document.getElementById('total-users').textContent = stats.total_users;
    document.getElementById('total-posts').textContent = stats.total_posts;
    document.getElementById('total-comments').textContent = stats.total_comments;
    document.getElementById('active-users').textContent = stats.active_users_24h;
}

// Массив пользователей для демонстрации
let usersData = [];

// Загрузка пользователей из API
// Загрузка пользователей из API
async function loadUsersList(filters = {}, sortBy = null) {
    try {
        // Получение пользователей с API
        const users = await AdminAPI.getUsers();
        usersData = users; // Сохраняем данные в глобальной переменной

        // Применяем фильтры
        const filteredUsers = users.filter(user => {
            const matchesUsername = filters.username ? user.username.includes(filters.username) : true;
            const matchesEmail = filters.email ? user.email.includes(filters.email) : true;
            const matchesRole = filters.role ? user.role === filters.role : true;
            return matchesUsername && matchesEmail && matchesRole;
        });

        // Применяем сортировку
        if (sortBy) {
            filteredUsers.sort((a, b) => {
                if (sortBy === 'username') {
                    return a.username.localeCompare(b.username);
                } else if (sortBy === 'role') {
                    return (a.role || '').localeCompare(b.role || '');
                }
                return 0;
            });
        }

        // Обновляем интерфейс
        const usersList = document.getElementById('users-list');
        usersList.innerHTML = filteredUsers.map(user => `
        <div class="user-item">
        <span>${user.username} (${user.email})</span>
        <button onclick="editUser(${user.id})" class="edit-btn">
            <i class="fas fa-edit"></i> 
        </button>
        <button onclick="deleteUser(${user.id})" class="delete-btn">
            <i class="fas fa-trash-alt"></i> 
        </button>
    </div>
    
        `).join('');
    } catch (error) {
        console.error('Error loading users:', error);
        showError('Failed to load users list');
    }
}


// Обработчик кнопки фильтрации
document.getElementById('apply-filters').addEventListener('click', () => {
    const usernameFilter = document.getElementById('username-filter').value.trim();
    const emailFilter = document.getElementById('email-filter').value.trim();
    const roleFilter = document.getElementById('role-filter').value;

    const filters = {
        username: usernameFilter,
        email: emailFilter,
        role: roleFilter
    };

    loadUsersList(filters); // Загрузка с фильтрами
});

// Обработчик кнопки сброса фильтров
document.getElementById('reset-filters').addEventListener('click', () => {
    document.getElementById('username-filter').value = '';
    document.getElementById('email-filter').value = '';
    document.getElementById('role-filter').value = '';

    loadUsersList(); // Загрузка без фильтров
});

// Сортировка пользователей
let currentSortBy = null;
document.getElementById('users-list').addEventListener('click', (event) => {
    if (event.target.dataset.sortBy) {
        currentSortBy = event.target.dataset.sortBy;
        loadUsersList({}, currentSortBy); // Загрузка с сортировкой
    }
});

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', () => {
    loadUsersList(); // Загрузка без фильтров и сортировки
});




async function editUser(userId) {
    const username = prompt('Enter new username:');
    const email = prompt('Enter new email:');
    if (username && email) {
        try {
            await AdminAPI.editUser(userId, username, email);
            alert('User updated successfully');
            loadUsersList();
        } catch (error) {
            console.error('Error editing user:', error);
            showError('Failed to update user');
        }
    }
}


function cancelEdit(userId) {
    loadUsersList();
}


async function handleAdminLogin(event) {
    event.preventDefault();

    const email = document.getElementById('admin-email').value;
    const password = document.getElementById('admin-password').value;

    try {
        const response = await fetch('/api/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ email, password })
        });

        if (!response.ok) {
            throw new Error('Login failed');
        }

        const { token } = await response.json();
        AdminAuth.setToken(token);
        showAdminContent();
    } catch (error) {
        console.error('Login error:', error);
        showError('Login failed. Please check your credentials.');
    }
}

// UI Management: Функция для обработки отправки рассылки
async function handleBroadcastEmail(event) {
    event.preventDefault();

    const recipient = document.getElementById('email-recipient').value;
    const subject = document.getElementById('email-subject').value;
    const body = document.getElementById('email-body').value;

    if (!recipient || !subject || !body) {
        showError('Please fill in all fields.');
        return;
    }

    try {
        const response = await AdminAPI.sendBroadcastEmail(recipient, subject, body);

        if (response.ok) {
            const text = await response.text();
            console.log('Raw server response:', text);

            try {
                const data = JSON.parse(text);
                if (data.success) {
                    showSuccess('Email sent successfully');
                } else {
                    showError(data.message || 'Failed to send email');
                }
            } catch (jsonError) {
                console.error('Error parsing JSON:', jsonError);
                showSuccess('Email sent successfully');
            }
        } else {
            showError('Failed to send email. Server returned an error.');
        }

        document.getElementById('broadcast-form').reset();
    } catch (error) {
        console.error('Error sending broadcast:', error);
        showError(error.message || 'Failed to send email');
    }
}
function showWarning(message) {
    // Implementation to display warning message
    console.warn(message);
    alert(message);
}


async function deleteUser(userId) {
    if (!confirm('Are you sure you want to delete this user?')) {
        return;
    }

    try {
        await AdminAPI.deleteUser(userId);
        showSuccess('User deleted successfully');
        loadUsersList();
    } catch (error) {
        console.error('Error deleting user:', error);
        showError('Failed to delete user');
    }
}

function handleLogout() {
    AdminAuth.removeToken();
    showLoginForm();
}
function showNotification(message, type) {
    const notification = document.getElementById('notification');
    notification.textContent = message;
    notification.className = `notification ${type}`;
    setTimeout(() => notification.className = 'notification', 5000);
}


function showWarning(message) {
    showNotification(message, 'warning');
}

function showError(message) {
    const notification = document.getElementById('notification');
    notification.textContent = message;
    notification.className = 'notification error';
    setTimeout(() => notification.className = 'notification', 3000);
}

function showSuccess(message) {
    const notification = document.getElementById('notification');
    notification.textContent = message;
    notification.className = 'notification success';
    setTimeout(() => notification.className = 'notification', 3000);
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    if (AdminAuth.isAuthenticated()) {
        showAdminContent();
    } else {
        showLoginForm();
    }

    document.getElementById('admin-login-form').addEventListener('submit', handleAdminLogin);
    document.getElementById('logout-btn').addEventListener('click', handleLogout);

    const applyFiltersBtn = document.getElementById('apply-filters');
    const resetFiltersBtn = document.getElementById('reset-filters');

    applyFiltersBtn.addEventListener('click', () => {
        const filters = {
            username: document.getElementById('username-filter').value,
            email: document.getElementById('email-filter').value,
            role: document.getElementById('role-filter').value,
        };
        loadUsersList(filters);
    });

    resetFiltersBtn.addEventListener('click', () => {
        document.getElementById('username-filter').value = '';
        document.getElementById('email-filter').value = '';
        document.getElementById('role-filter').value = '';
        loadUsersList();
    });
});