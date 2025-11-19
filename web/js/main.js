class App {
    constructor() {
        this.init();
    }

    init() {
        this.setupNavbar();
        this.updateNavbar();
        this.setupMobileMenu();
    }

    setupNavbar() {
        this.updateNavbar();
    }

    updateNavbar() {
        const navbarMenu = document.getElementById('navbarMenu');
        if (!navbarMenu) return;

        const isAuthenticated = api.isAuthenticated();
        const user = api.getUser();

        if (isAuthenticated && user) {
            navbarMenu.innerHTML = `
                <a href="products.html" class="nav-link">
                    <i class="fas fa-shopping-bag"></i>
                    Магазин
                </a>
                <a href="cart.html" class="nav-link cart-link">
                    <i class="fas fa-shopping-cart"></i>
                    Корзина
                    <span class="cart-badge" id="cartBadge" style="display: none;">0</span>
                </a>
                <div class="user-menu">
                    <a href="profile.html" class="nav-link">
                        <i class="fas fa-user"></i>
                        ${user.first_name}
                    </a>
                    <button onclick="app.logout()" class="nav-link logout-btn">
                        <i class="fas fa-sign-out-alt"></i>
                        Выйти
                    </button>
                </div>
            `;
        } else {
            navbarMenu.innerHTML = `
                <a href="products.html" class="nav-link">
                    <i class="fas fa-shopping-bag"></i>
                    Магазин
                </a>
                <div class="auth-links">
                    <a href="login.html" class="nav-link">
                        <i class="fas fa-sign-in-alt"></i>
                        Войти
                    </a>
                    <a href="register.html" class="nav-link register-btn">
                        Регистрация
                    </a>
                </div>
            `;
        }

        // Load cart count if authenticated
        if (isAuthenticated) {
            this.loadCartCount();
        }
    }

    setupMobileMenu() {
        const toggle = document.getElementById('navbarToggle');
        const menu = document.getElementById('navbarMenu');

        if (toggle && menu) {
            toggle.addEventListener('click', () => {
                menu.classList.toggle('active');
            });

            // Close menu when clicking outside
            document.addEventListener('click', (e) => {
                if (!navbar.contains(e.target)) {
                    menu.classList.remove('active');
                }
            });
        }
    }

    async loadCartCount() {
        try {
            const cartData = await api.getCart();
            const badge = document.getElementById('cartBadge');
            
            if (badge && cartData.item_count > 0) {
                badge.textContent = cartData.item_count;
                badge.style.display = 'flex';
            }
        } catch (error) {
            console.error('Failed to load cart count:', error);
        }
    }

    logout() {
        if (confirm('Вы уверены, что хотите выйти?')) {
            api.removeToken();
            window.location.href = 'index.html';
        }
    }

    showNotification(message, type = 'success') {
        // Create notification element
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.innerHTML = `
            <i class="fas fa-${type === 'success' ? 'check' : 'exclamation'}-circle"></i>
            <span>${message}</span>
        `;

        // Add styles
        notification.style.cssText = `
            position: fixed;
            top: 100px;
            right: 20px;
            background: ${type === 'success' ? '#d4edda' : '#f8d7da'};
            color: ${type === 'success' ? '#155724' : '#721c24'};
            padding: 1rem 1.5rem;
            border-radius: 8px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
            z-index: 1001;
            display: flex;
            align-items: center;
            gap: 0.5rem;
            max-width: 400px;
            animation: slideIn 0.3s ease;
        `;

        document.body.appendChild(notification);

        // Remove after 5 seconds
        setTimeout(() => {
            notification.style.animation = 'slideOut 0.3s ease';
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.parentNode.removeChild(notification);
                }
            }, 300);
        }, 5000);
    }
    loadProductsManager() {
    if (window.products && typeof window.products.updateNavbar === 'function') {
        window.products.updateNavbar();
        }
    }
}

// Add CSS for notifications
const notificationStyles = `
@keyframes slideIn {
    from { transform: translateX(100%); opacity: 0; }
    to { transform: translateX(0); opacity: 1; }
}

@keyframes slideOut {
    from { transform: translateX(0); opacity: 1; }
    to { transform: translateX(100%); opacity: 0; }
}
`;

const styleSheet = document.createElement('style');
styleSheet.textContent = notificationStyles;
document.head.appendChild(styleSheet);

// Global app instance
const app = new App();

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.app = new App();
});