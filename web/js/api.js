class API {
    constructor() {
        this.baseURL = 'http://localhost:3000/api/v1';
        this.token = localStorage.getItem('auth_token');
    }

    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        };

        if (this.token) {
            config.headers.Authorization = `Bearer ${this.token}`;
        }

        try {
            const response = await fetch(url, config);
            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || 'Request failed');
            }

            return data;
        } catch (error) {
            if (error.message === 'Failed to fetch') {
                throw new Error('Network error. Please check your connection.');
            }
            throw error;
        }
    }

    // Auth methods
    async login(credentials) {
        return this.request('/auth/login', {
            method: 'POST',
            body: JSON.stringify(credentials)
        });
    }

    async register(userData) {
        return this.request('/auth/register', {
            method: 'POST',
            body: JSON.stringify(userData)
        });
    }

    async getProfile() {
        return this.request('/user/profile');
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
            body: JSON.stringify(data)
        });
    }

    async getUserOrders() {
        return this.request('/orders');
    }

    setToken(token) {
        this.token = token;
        localStorage.setItem('auth_token', token);
    }

    removeToken() {
        this.token = null;
        localStorage.removeItem('auth_token');
        localStorage.removeItem('user');
    }

    isAuthenticated() {
        return !!this.token;
    }

    getUser() {
        const user = localStorage.getItem('user');
        return user ? JSON.parse(user) : null;
    }

    setUser(user) {
        localStorage.setItem('user', JSON.stringify(user));
    }
}

// Global API instance
const api = new API();