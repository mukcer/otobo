class AuthManager {
    constructor() {
        this.init();
    }

    init() {
        this.setupLoginForm();
        this.setupRegisterForm();
        this.checkAuthentication();
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

    async handleLogin(e) {
        e.preventDefault();
        
        const form = e.target;
        const button = form.querySelector('button[type="submit"]');
        const errorAlert = document.getElementById('errorAlert');
        const errorText = document.getElementById('errorText');
        
        // Clear previous errors
        this.clearErrors();
        this.hideAlert(errorAlert);

        // Validate form
        if (!this.validateLoginForm(form)) {
            return;
        }

        // Get form data
        const formData = new FormData(form);
        const credentials = {
            email: formData.get('email'),
            password: formData.get('password')
        };

        // Show loading state
        this.setLoading(button, true);

        try {
            const response = await api.login(credentials);
            
            // Save auth data
            api.setToken(response.token);
            api.setUser(response.user);
            
            // Redirect to home page
            window.location.href = 'index.html';
            
        } catch (error) {
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
        
        // Clear previous errors
        this.clearErrors();
        this.hideAlert(errorAlert);
        this.hideAlert(successAlert);

        // Validate form
        if (!this.validateRegisterForm(form)) {
            return;
        }

        // Get form data
        const formData = new FormData(form);
        const userData = {
            first_name: formData.get('first_name'),
            last_name: formData.get('last_name'),
            email: formData.get('email'),
            phone: formData.get('phone') || '',
            password: formData.get('password')
        };

        // Show loading state
        this.setLoading(button, true);

        try {
            await api.register(userData);
            
            // Show success message
            this.showAlert(successAlert);
            
            // Auto login after registration
            setTimeout(async () => {
                try {
                    const loginResponse = await api.login({
                        email: userData.email,
                        password: userData.password
                    });
                    
                    api.setToken(loginResponse.token);
                    api.setUser(loginResponse.user);
                    
                    window.location.href = 'index.html';
                    
                } catch (loginError) {
                    // Redirect to login page if auto-login fails
                    window.location.href = 'login';
                }
            }, 2000);
            
        } catch (error) {
            this.showError(errorAlert, errorText, error.message);
        } finally {
            this.setLoading(button, false);
        }
    }

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
            button.disabled = true;
            button.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Загрузка...';
            button.classList.add('loading');
        } else {
            button.disabled = false;
            if (button.id === 'loginBtn') {
                button.innerHTML = '<i class="fas fa-sign-in-alt"></i> Войти';
            } else {
                button.innerHTML = '<i class="fas fa-user-plus"></i> Зарегистрироваться';
            }
            button.classList.remove('loading');
        }
    }

    checkAuthentication() {
        if (api.isAuthenticated() && (window.location.pathname.includes('login') || window.location.pathname.includes('register'))) {
            window.location.href = '/';
        }
    }
}

// Initialize auth manager when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new AuthManager();
});