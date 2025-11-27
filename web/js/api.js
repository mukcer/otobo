 // api.js
class Api {
    constructor() {
        this.baseURL = '/api/v1'; //window.location.origin; // или ваш API URL
        this.token = null;
        this.user = null;
    }

    // 🔐 Установка токена
    setToken(token) {
        this.token = token;
        // Также сохраняем в localStorage для persistence
        if (token) {
            localStorage.setItem('auth_token', token);
        }
    }

    // 👤 Установка пользователя
    setUser(user) {
        this.user = user;
    }

    // 🗑️ Удаление токена
    removeToken() {
        this.token = null;
        localStorage.removeItem('auth_token');
    }

      // ✅ Проверка аутентификации
    isAuthenticated() {
        return !!(this.token || localStorage.getItem('auth_token'));
    }

    // 👤 Получение пользователя
    getUser() {
        return this.user;
    }

    // 🗑️ Удаление пользователя
    removeUser() {
        this.user = null;
    }

    // 🌐 Базовый метод для HTTP запросов
    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers,
            },
            ...options,
        };

        // Добавляем токен авторизации если есть
        const token = this.token || localStorage.getItem('auth_token');
        if (token) {
            config.headers['Authorization'] = `Bearer ${token}`;
        }
        if (options.body) {
            config.body = JSON.stringify(options.body);
        }
       try {
        const response = await fetch(url, config);
        
        if (!response.ok) {
            let errorData;
            try {
                errorData = await response.json();
            } catch {
                errorData = { error: `HTTP error! status: ${response.status}` };
            }
            throw new Error(errorData.error || errorData.message || `Request failed with status ${response.status}`);
        }

        return await response.json();
        
    } catch (error) {
        console.error('API request failed:', error);
        throw error;
    }
    }

    // 🔐 Логин
    async login(credentials) {
        return await this.request('/auth/login', {
            method: 'POST',
            body: credentials,
        });
    }

    // 📝 Регистрация
    async register(userData) {
        return await this.request('/auth/register', {
            method: 'POST',
            body: userData,
        });
    }

    // 🔄 Синхронизация
    async get(endpoint, options = {}) {
        return await this.request(endpoint, {
            method: 'GET',
            ...options,
        });
    }

    // 📤 POST запрос
    async post(endpoint, data) {
        return await this.request(endpoint, {
            method: 'POST',
            body: data,
        });
    }

    // 🗑️ DELETE запрос
    async delete(endpoint) {
        return await this.request(endpoint, {
            method: 'DELETE',
        });
    }

    // 👤 Получение профиля
    async getProfile() {
        return await this.get('/user/profile');
    }

    // 🔄 Синхронизация данных
    async sync(lastSync = null) {
        const headers = {};
        if (lastSync) {
            headers['X-Last-Sync'] = lastSync;
        }
        
        return await this.get('/auth/sync', { headers });
    }

    // 🚪 Выход
    async logout() {
        return await this.post('/auth/logout');
    }


    // Products methods
    async getProducts(params = {}) {
        const queryString = new URLSearchParams(params).toString();
        return this.request(`/products?${queryString}`);
    }

    async getProduct(slug) {
        return this.request(`/products/${slug}`);
    }
    async getProductByID(id) {
        return this.request(`/products/id/${id}`);
    }

    async getCategories() {
        return this.request('/products/categories');
    }
    async getColors() {
        return this.request('/colors');
    }        

    // Cart methods
    async getCart() {
        return this.request('/cart');
    }

    async addToCart(data) {
        return this.request('/cart', {
            method: 'POST',
            body: JSON.stringify(data)
        });
    }

    async updateCartItem(id, data) {
        return this.request(`/cart/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data)
        });
    }

    async removeFromCart(id) {
        return this.request(`/cart/${id}`, {
            method: 'DELETE'
        });
    }

    async clearCart() {
        return this.request('/cart', {
            method: 'DELETE'
        });
    }

    // Orders methods
    async createOrder(data) {
        return this.request('/orders', {
            method: 'POST',
            body: data
        });
    }

    async getUserOrders() {
        return this.request('/orders');
    }

}
// Создаем глобальный экземпляр API
const api = new Api();
