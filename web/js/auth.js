class AuthManager {
    constructor() {
        this.STORAGE_KEYS = {
            TOKEN: 'auth_token',
            USER: 'user_data',
            SESSION_SYNC: 'session_synced',
            LAST_SYNC: 'last_sync_time'
        };
        this.init();
    }

    init() {
        this.setupLoginForm();
        this.setupRegisterForm();
        this.checkAuthentication();
        this.setupAutoSync();
    }
    setupLoginForm() {
        const loginForm = document.getElementById('loginForm');
        if (loginForm) {
            loginForm.addEventListener('submit', this.handleLogin.bind(this));
        }
    }

    setupRegisterForm() {
        const registerForm = document.getElementById('registerForm');
        if (registerForm) {
            registerForm.addEventListener('submit', this.handleRegister.bind(this));
        }
    }

    // 🔄 Настройка автоматической синхронизации
    setupAutoSync() {
        // Синхронизация при загрузке страницы
        this.syncWithServer();
        
        // Периодическая синхронизация каждые 5 минут
        setInterval(() => this.syncWithServer(), 5 * 60 * 1000);
        
        // Синхронизация при возвращении на страницу
        document.addEventListener('visibilitychange', () => {
            if (!document.hidden) {
                this.syncWithServer();
            }
        });
    }

    // 🔄 Синхронизация данных между localStorage и Redis
    async syncWithServer() {
        if (!api.isAuthenticated()) return;

        try {
            const lastSync = localStorage.getItem(this.STORAGE_KEYS.LAST_SYNC);
            const now = Date.now();
            
            // Синхронизируем не чаще чем раз в 30 секунд
            if (lastSync && (now - parseInt(lastSync)) < 30000) {
                return;
            }

            const response = await api.get('/api/auth/sync', {
                headers: {
                    'X-Last-Sync': lastSync || '0'
                }
            });

            if (response.user) {
                // Обновляем локальные данные с сервера
                this.saveUserData(response.user);
            }

            localStorage.setItem(this.STORAGE_KEYS.LAST_SYNC, now.toString());
            localStorage.setItem(this.STORAGE_KEYS.SESSION_SYNC, 'true');
            
        } catch (error) {
            console.log('Sync failed, using local data:', error.message);
        }
    }

    // 💾 Сохранение пользовательских данных
    saveUserData(user) {
        // Сохраняем в localStorage для мгновенного доступа
        localStorage.setItem(this.STORAGE_KEYS.USER, JSON.stringify(user));
        
        // Сохраняем в памяти приложения
        api.setUser(user);
    }

    // 🔍 Проверка аутентификации с приоритетом localStorage
    checkAuthentication() {
        const token = localStorage.getItem(this.STORAGE_KEYS.TOKEN);
        const userData = localStorage.getItem(this.STORAGE_KEYS.USER);

        // Если есть данные в localStorage, используем их мгновенно
        if (token && userData) {
            api.setToken(token);
            api.setUser(JSON.parse(userData));
            
            // Обновляем навигацию сразу
            if (window.app && typeof window.app.updateNavbar === 'function') {
                window.app.updateNavbar();
            }

            // Если на странице логина/регистрации - редирект
            if ((window.location.pathname.includes('login') || 
                 window.location.pathname.includes('register')) && 
                api.isAuthenticated()) {
                window.location.href = '/products';
            }

            // Фоновая синхронизация с сервером
            this.syncWithServer();
        } else if (api.isAuthenticated() && 
                  (window.location.pathname.includes('login') || 
                   window.location.pathname.includes('register'))) {
            window.location.href = '/products';
        }
    }

    async handleLogin(e) {
        e.preventDefault();
        
        const form = e.target;
        const button = form.querySelector('button[type="submit"]');
        const errorAlert = document.getElementById('errorAlert');
        const errorText = document.getElementById('errorText');
        
        this.clearErrors();
        this.hideAlert(errorAlert);

        if (!this.validateLoginForm(form)) {
            return;
        }

        const formData = new FormData(form);
        const credentials = {
            email: formData.get('email'),
            password: formData.get('password'),
            // Добавляем информацию о клиенте для синхронизации
            client_timestamp: Date.now(),
            has_local_data: !!localStorage.getItem(this.STORAGE_KEYS.USER)
        };

        this.setLoading(button, true);

        try {
            const response = await api.login(credentials);
            
            // ✅ Сохраняем данные в localStorage
            localStorage.setItem(this.STORAGE_KEYS.TOKEN, response.token);
            this.saveUserData(response.user);
            
            // ✅ Сохраняем в Redis через API (сессия создается на сервере)
            await this.createServerSession(response.user);
            
            // Обновляем навигацию
            if (window.app && typeof window.app.updateNavbar === 'function') {
                window.app.updateNavbar();
            }
            
            // Редирект на главную
            window.location.href = '/';
            
        } catch (error) {
            // ❌ При ошибке очищаем невалидные данные
            this.clearAuthData();
            this.showError(errorAlert, errorText, error.message);
        } finally {
            this.setLoading(button, false);
        }
    }

    async handleRegister(e) {
        e.preventDefault();
        
        const form = e.target;
        const button = form.querySelector('button[type="submit"]');
        const errorAlert = document.getElementById('errorAlert');
        const errorText = document.getElementById('errorText');
        const successAlert = document.getElementById('successAlert');
        
        this.clearErrors();
        this.hideAlert(errorAlert);
        this.hideAlert(successAlert);

        if (!this.validateRegisterForm(form)) {
            return;
        }

        const formData = new FormData(form);
        const userData = {
            first_name: formData.get('first_name'),
            last_name: formData.get('last_name'),
            email: formData.get('email'),
            phone: formData.get('phone') || '',
            password: formData.get('password'),
            client_timestamp: Date.now()
        };

        this.setLoading(button, true);

        try {
            await api.register(userData);
            
            this.showAlert(successAlert);
            
            // Автоматический логин после регистрации
            setTimeout(async () => {
                try {
                    const loginResponse = await api.login({
                        email: userData.email,
                        password: userData.password
                    });
                    
                    // ✅ Сохраняем в localStorage и Redis
                    localStorage.setItem(this.STORAGE_KEYS.TOKEN, loginResponse.token);
                    this.saveUserData(loginResponse.user);
                    await this.createServerSession(loginResponse.user);
                    
                    window.location.href = 'index.html';
                    
                } catch (loginError) {
                    // Редирект на логин если авто-логин не удался
                    window.location.href = 'login';
                }
            }, 2000);
            
        } catch (error) {
            this.showError(errorAlert, errorText, error.message);
        } finally {
            this.setLoading(button, false);
        }
    }

    // 🔐 Создание сессии на сервере (Redis)
    async createServerSession(user) {
        try {
            await api.post('/api/auth/session', {
                user_id: user.id,
                login_time: new Date().toISOString(),
                user_agent: navigator.userAgent,
                client_data: {
                    last_sync: localStorage.getItem(this.STORAGE_KEYS.LAST_SYNC),
                    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone
                }
            });
        } catch (error) {
            console.warn('Server session creation failed:', error);
            // Не прерываем процесс, т.к. локальные данные уже сохранены
        }
    }

    // 🚪 Выход с очисткой всех данных
    async logout() {
        if (!confirm('Вы уверены, что хотите выйти?')) {
            return;
        }

        try {
            // Отправляем запрос на сервер для очистки Redis сессии
            await api.post('/api/auth/logout');
        } catch (error) {
            console.warn('Server logout failed:', error);
        } finally {
            // Всегда очищаем локальные данные
            this.clearAuthData();
            
            // Обновляем интерфейс
            if (window.app && typeof window.app.updateNavbar === 'function') {
                window.app.updateNavbar();
            }
            
            // Редирект
            window.location.href = '/login';
        }
    }

    // 🗑️ Очистка всех аутентификационных данных
    clearAuthData() {
        // Очищаем localStorage
        localStorage.removeItem(this.STORAGE_KEYS.TOKEN);
        localStorage.removeItem(this.STORAGE_KEYS.USER);
        localStorage.removeItem(this.STORAGE_KEYS.SESSION_SYNC);
        
        // Очищаем память приложения
        api.removeToken();
        api.removeUser();
    }

    // 📱 Получение пользовательских данных с приоритетом localStorage
    getUserData() {
        // Сначала пробуем из localStorage (самый быстрый)
        const localUser = localStorage.getItem(this.STORAGE_KEYS.USER);
        if (localUser) {
            return JSON.parse(localUser);
        }
        
        // Затем из памяти приложения
        return api.getUser();
    }

    // Остальные методы остаются без изменений...
    validateLoginForm(form) {
        let isValid = true;
        const email = form.email.value.trim();
        const password = form.password.value;

        if (!email) {
            this.showFieldError('email', 'Email обязателен');
            isValid = false;
        } else if (!this.isValidEmail(email)) {
            this.showFieldError('email', 'Некорректный email адрес');
            isValid = false;
        }

        if (!password) {
            this.showFieldError('password', 'Пароль обязателен');
            isValid = false;
        } else if (password.length < 6) {
            this.showFieldError('password', 'Пароль должен содержать минимум 6 символов');
            isValid = false;
        }

        return isValid;
    }

    validateRegisterForm(form) {
        let isValid = true;
        const firstName = form.first_name.value.trim();
        const lastName = form.last_name.value.trim();
        const email = form.email.value.trim();
        const password = form.password.value;
        const confirmPassword = form.confirmPassword.value;

        if (!firstName) {
            this.showFieldError('firstName', 'Имя обязательно');
            isValid = false;
        }

        if (!lastName) {
            this.showFieldError('lastName', 'Фамилия обязательна');
            isValid = false;
        }

        if (!email) {
            this.showFieldError('email', 'Email обязателен');
            isValid = false;
        } else if (!this.isValidEmail(email)) {
            this.showFieldError('email', 'Некорректный email адрес');
            isValid = false;
        }

        if (!password) {
            this.showFieldError('password', 'Пароль обязателен');
            isValid = false;
        } else if (password.length < 6) {
            this.showFieldError('password', 'Пароль должен содержать минимум 6 символов');
            isValid = false;
        }

        if (!confirmPassword) {
            this.showFieldError('confirmPassword', 'Подтверждение пароля обязательно');
            isValid = false;
        } else if (password !== confirmPassword) {
            this.showFieldError('confirmPassword', 'Пароли не совпадают');
            isValid = false;
        }

        return isValid;
    }

    isValidEmail(email) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
    }

    showFieldError(fieldName, message) {
        const errorElement = document.getElementById(`${fieldName}Error`);
        const inputElement = document.getElementById(fieldName);
        
        if (errorElement && inputElement) {
            errorElement.textContent = message;
            inputElement.classList.add('error');
        }
    }

    clearErrors() {
        const errorMessages = document.querySelectorAll('.error-message');
        const errorInputs = document.querySelectorAll('.error');
        
        errorMessages.forEach(el => el.textContent = '');
        errorInputs.forEach(el => el.classList.remove('error'));
    }

    showAlert(alertElement) {
        if (alertElement) {
            alertElement.style.display = 'flex';
        }
    }

    hideAlert(alertElement) {
        if (alertElement) {
            alertElement.style.display = 'none';
        }
    }

    showError(alertElement, textElement, message) {
        if (alertElement && textElement) {
            textElement.textContent = message;
            alertElement.style.display = 'flex';
        }
    }

    setLoading(button, isLoading) {
        if (isLoading) {
            button.dataset.originalHTML = button.innerHTML;
            button.disabled = true;
            button.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Загрузка...';
            button.classList.add('loading');
        } else {
            button.disabled = false;
            button.innerHTML = button.dataset.originalHTML || button.innerHTML;
            button.classList.remove('loading');
        }
    }
}

// Инициализация когда DOM загружен
document.addEventListener('DOMContentLoaded', () => {
    new AuthManager();
});