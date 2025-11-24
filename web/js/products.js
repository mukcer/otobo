class ProductsManager {
    constructor() {
        this.currentPage = 1;
        this.limit = 12;
        this.filters = {
            category: '',
            size: [],
            color: [],
            minPrice: '',
            maxPrice: '',
            inStock: false
        };
        this.sortBy = 'created_at';
        this.sortOrder = 'desc';
        this.viewMode = 'grid'; // 'grid' or 'list'
        
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.loadCategories();
        this.loadColors();
        this.loadProducts();
        this.updateNavbar();
    }

    setupEventListeners() {
        // Filter events
        document.getElementById('applyFilters').addEventListener('click', () => this.applyFilters());
        document.getElementById('clearFilters').addEventListener('click', () => this.clearFilters());
        document.getElementById('inStockOnly').addEventListener('change', (e) => {
            this.filters.inStock = e.target.checked;
        });

        // Sort events
        document.getElementById('sortBy').addEventListener('change', (e) => {
            this.sortBy = e.target.value;
            this.loadProducts();
        });

        document.getElementById('sortOrder').addEventListener('change', (e) => {
            this.sortOrder = e.target.value;
            this.loadProducts();
        });

        // View mode toggle
        document.getElementById('viewToggle').addEventListener('click', () => this.toggleViewMode());

        // Price filter events
        document.getElementById('minPrice').addEventListener('input', (e) => {
            this.filters.minPrice = e.target.value;
        });

        document.getElementById('maxPrice').addEventListener('input', (e) => {
            this.filters.maxPrice = e.target.value;
        });

        // Size filter events
        document.querySelectorAll('input[name="size"]').forEach(checkbox => {
            checkbox.addEventListener('change', (e) => {
                if (e.target.checked) {
                    this.filters.size.push(e.target.value);
                } else {
                    this.filters.size = this.filters.size.filter(size => size !== e.target.value);
                }
            });
        });

        // Modal events
        document.getElementById('closeModal').addEventListener('click', () => this.closeModal());
        document.getElementById('productModal').addEventListener('click', (e) => {
            if (e.target.id === 'productModal') this.closeModal();
        });

        // Mobile filters
        document.getElementById('mobileFiltersToggle').addEventListener('click', () => this.toggleMobileFilters());

        // Close mobile filters when clicking outside
        document.addEventListener('click', (e) => {
            const filtersSidebar = document.getElementById('filtersSidebar');
            const mobileToggle = document.getElementById('mobileFiltersToggle');
            
            if (!filtersSidebar.contains(e.target) && !mobileToggle.contains(e.target)) {
                filtersSidebar.classList.remove('active');
            }
        });
    }

    async loadCategories() {
        try {
            const data = await api.getCategories();
            const container = document.getElementById('categoriesFilter');
            
            container.innerHTML = data.map(category => `
                <label class="filter-checkbox">
                    <input type="radio" name="category" value="${category.slug}">
                    <span class="checkmark"></span>
                    ${category.name}
                </label>
            `).join('');

            // Add category filter events
            container.querySelectorAll('input[name="category"]').forEach(radio => {
                radio.addEventListener('change', (e) => {
                    this.filters.category = e.target.value;
                });
            });

        } catch (error) {
            console.error('Failed to load categories:', error);
        }
    }

async loadColors() {
    try {
        const data = await api.getColors();
        const container = document.getElementById('colorsFilter');
        container.innerHTML = '<div class="loading">Загрузка цветов...</div>';
        
        // В зависимости от структуры ответа API
        const colorsArray = data.colors || colors; // если colors вложен в объект
        
        if (!colorsArray || colorsArray.length === 0) {
            container.innerHTML = '<div class="no-colors">Цвета не найдены</div>';
            return;
        }

        this.renderColors(colorsArray, container);
        this.attachColorFilterEvents(container);

    } catch (error) {
        console.error('Failed to load colors:', error);
        this.handleColorLoadError(error);
    }
}

renderColors(colors, container) {
    container.innerHTML = colors.map(color => `
        <div class="color-option ${color.available ? '' : 'disabled'}" 
             style="background-color: ${color.value}" 
             title="${color.name}${!color.available ? ' (недоступен)' : ''}"
             data-color="${color.value}"
             ${!color.available ? 'disabled' : ''}>
        </div>
    `).join('');
}

attachColorFilterEvents(container) {
    container.querySelectorAll('.color-option:not(.disabled)').forEach(colorOption => {
        colorOption.addEventListener('click', (e) => {
            const color = e.target.dataset.color;
            e.target.classList.toggle('selected');
            
            if (e.target.classList.contains('selected')) {
                this.filters.color.push(color);
            } else {
                this.filters.color = this.filters.color.filter(c => c !== color);
            }
        });
    });
}

handleColorLoadError(error) {
    const container = document.getElementById('colorsFilter');
    
    if (error.message.includes('network') || error.message.includes('Failed to fetch')) {
        container.innerHTML = `
            <div class="error">
                Ошибка сети. Проверьте подключение к интернету.
                <button onclick="this.loadColors()">Повторить</button>
            </div>
        `;
    } else {
        container.innerHTML = `
            <div class="error">
                Не удалось загрузить цвета
                <button onclick="this.loadDemoColors()">Использовать демо-данные</button>
            </div>
        `;
    }
}

// Дополнительный метод для демо-данных (опционально)
loadDemoColors() {
    const colors = [
        { name: 'Черный', value: '#000000', available: true },
        { name: 'Белый', value: '#FFFFFF', available: true },
        { name: 'Красный', value: '#FF0000', available: true },
        { name: 'Синий', value: '#0000FF', available: true },
        { name: 'Зеленый', value: '#008000', available: true },
        { name: 'Розовый', value: '#FFC0CB', available: true },
        { name: 'Бежевый', value: '#F5F5DC', available: true },
        { name: 'Серый', value: '#808080', available: true }
    ];
    
    const container = document.getElementById('colorsFilter');
    this.renderColors(colors, container);
    this.attachColorFilterEvents(container);
}

    async loadProducts() {
        const grid = document.getElementById('productsGrid');
        grid.innerHTML = `
            <div class="loading-spinner">
                <i class="fas fa-spinner fa-spin"></i>
                <p>Загрузка товаров...</p>
            </div>
        `;

        try {
            const params = {
                page: this.currentPage,
                limit: this.limit,
                sort: this.sortBy,
                order: this.sortOrder
            };

            // Add filters to params
            if (this.filters.category) params.category = this.filters.category;
            if (this.filters.size.length > 0) params.size = this.filters.size.join(',');
            if (this.filters.color.length > 0) params.color = this.filters.color.join(',');
            if (this.filters.minPrice) params.min_price = this.filters.minPrice;
            if (this.filters.maxPrice) params.max_price = this.filters.maxPrice;
            if (this.filters.inStock) params.in_stock = true;

            const data = await api.getProducts(params);
            this.renderProducts(data.products);
            this.renderPagination(data.total, data.pages);
            this.updateProductsCount(data.total);

        } catch (error) {
            grid.innerHTML = `
                <div class="no-results">
                    <i class="fas fa-exclamation-triangle"></i>
                    <h3>Ошибка загрузки</h3>
                    <p>${error.message}</p>
                    <button onclick="products.loadProducts()" class="btn btn-primary">
                        <i class="fas fa-redo"></i>
                        Попробовать снова
                    </button>
                </div>
            `;
        }
    }

    renderProducts(products) {
        const grid = document.getElementById('productsGrid');
        
        if (products.length === 0) {
            grid.innerHTML = `
                <div class="no-results">
                    <i class="fas fa-search"></i>
                    <h3>Товары не найдены</h3>
                    <p>Попробуйте изменить параметры фильтрации</p>
                    <button onclick="products.clearFilters()" class="btn btn-primary">
                        <i class="fas fa-times"></i>
                        Очистить фильтры
                    </button>
                </div>
            `;
            return;
        }

        grid.innerHTML = products.map(product => `
            <div class="product-card" onclick="products.openProductModal('${product.slug}')">
                <div class="product-image">
                    ${product.images && product.images.length > 0 ? 
                        `<img src="${product.images[0]}" alt="${product.name}" loading="lazy">` :
                        `<i class="fas fa-image" style="font-size: 3rem; color: #ddd;"></i>`
                    }
                </div>
                <div class="product-info">
                    <div class="product-category">${product.category?.name || 'Без категории'}</div>
                    <h3 class="product-title">${product.name}</h3>
                    <p class="product-description">${product.description || ''}</p>
                    
                    <div class="product-price">
                        <span class="current-price">${this.formatPrice(product.price)} ₽</span>
                        ${product.compare_price ? 
                            `<span class="original-price">${this.formatPrice(product.compare_price)} ₽</span>` : 
                            ''
                        }
                    </div>
                    
                    <div class="product-meta">
                        <span class="${product.in_stock ? 'in-stock' : 'out-of-stock'}">
                            <i class="fas ${product.in_stock ? 'fa-check' : 'fa-times'}"></i>
                            ${product.in_stock ? 'В наличии' : 'Нет в наличии'}
                        </span>
                        <span class="product-sku">${product.sku}</span>
                    </div>
                </div>
            </div>
        `).join('');

        // Apply view mode
        grid.className = `products-grid ${this.viewMode}-view`;
    }

    renderPagination(total, totalPages) {
        const container = document.getElementById('pagination');
        
        if (totalPages <= 1) {
            container.innerHTML = '';
            return;
        }

        let paginationHTML = '';
        
        // Previous button
        paginationHTML += `
            <button onclick="products.previousPage()" ${this.currentPage === 1 ? 'disabled' : ''}>
                <i class="fas fa-chevron-left"></i>
            </button>
        `;

        // Page numbers
        for (let i = 1; i <= totalPages; i++) {
            if (i === 1 || i === totalPages || (i >= this.currentPage - 1 && i <= this.currentPage + 1)) {
                paginationHTML += `
                    <button onclick="products.goToPage(${i})" 
                            ${this.currentPage === i ? 'style="background: #e91e63; color: white;"' : ''}>
                        ${i}
                    </button>
                `;
            } else if (i === this.currentPage - 2 || i === this.currentPage + 2) {
                paginationHTML += `<span>...</span>`;
            }
        }

        // Next button
        paginationHTML += `
            <button onclick="products.nextPage()" ${this.currentPage === totalPages ? 'disabled' : ''}>
                <i class="fas fa-chevron-right"></i>
            </button>
        `;

        container.innerHTML = paginationHTML;
    }

    async openProductModal(slug) {
        try {
            const product = await api.getProduct(slug);
            this.renderProductModal(product);
        } catch (error) {
            console.error('Failed to load product:', error);
            app.showNotification('Ошибка загрузки товара', 'error');
        }
    }

    renderProductModal(product) {
        const modalBody = document.getElementById('modalBody');
        
        modalBody.innerHTML = `
            <div class="modal-product">
                <div class="product-gallery">
                    ${product.images && product.images.length > 0 ? 
                        product.images.map(img => `
                            <img src="${img}" alt="${product.name}" loading="lazy">
                        `).join('') :
                        `<div class="no-image">
                            <i class="fas fa-image"></i>
                            <p>Изображение отсутствует</p>
                         </div>`
                    }
                </div>
                
                <div class="product-details">
                    <h2>${product.name}</h2>
                    <div class="product-category">${product.category?.name || 'Без категории'}</div>
                    
                    <div class="product-price-large">
                        <span class="current-price">${this.formatPrice(product.price)} ₽</span>
                        ${product.compare_price ? 
                            `<span class="original-price">${this.formatPrice(product.compare_price)} ₽</span>` : 
                            ''
                        }
                    </div>
                    
                    <p class="product-description-full">${product.description || 'Описание отсутствует.'}</p>
                    
                    ${product.features && product.features.length > 0 ? `
                        <div class="product-features">
                            <h4>Особенности:</h4>
                            <ul>
                                ${product.features.map(feature => `<li>${feature}</li>`).join('')}
                            </ul>
                        </div>
                    ` : ''}
                    
                    <div class="product-variations">
                        <h4>Доступные варианты:</h4>
                        <div class="variations-list">
                            ${product.variations && product.variations.length > 0 ? 
                                product.variations.map(variation => `
                                    <div class="variation-item ${variation.quantity === 0 ? 'out-of-stock' : ''}">
                                        <span class="variation-size">${variation.size?.name}</span>
                                        <span class="variation-color" style="background-color: ${variation.color?.value}"></span>
                                        <span class="variation-quantity">${variation.quantity} шт.</span>
                                    </div>
                                `).join('') :
                                '<p>Варианты отсутствуют</p>'
                            }
                        </div>
                    </div>
                    
                    <div class="product-actions">
                        <button class="btn btn-primary" onclick="products.addToCart(${product.id})" 
                                ${!product.in_stock ? 'disabled' : ''}>
                            <i class="fas fa-shopping-cart"></i>
                            Добавить в корзину
                        </button>
                        <button class="btn btn-secondary" onclick="products.closeModal()">
                            Закрыть
                        </button>
                    </div>
                </div>
            </div>
        `;

        document.getElementById('productModal').style.display = 'block';
    }

    closeModal() {
        document.getElementById('productModal').style.display = 'none';
    }

    async addToCart(productId) {
        if (!api.isAuthenticated()) {
            app.showNotification('Для добавления в корзину необходимо войти в систему', 'error');
            window.location.href = '/login';
            return;
        }

        try {
            // For demo - in real app, you'd select variation
            await api.addToCart({
                product_id: productId,
                variation_id: 1, // First variation for demo
                quantity: 1
            });

            app.showNotification('Товар добавлен в корзину', 'success');
            this.closeModal();
            this.updateCartCount();
            
        } catch (error) {
            app.showNotification(error.message, 'error');
        }
    }

    applyFilters() {
        this.currentPage = 1;
        this.loadProducts();
        this.closeMobileFilters();
    }

    clearFilters() {
        this.filters = {
            category: '',
            size: [],
            color: [],
            minPrice: '',
            maxPrice: '',
            inStock: false
        };

        // Reset UI
        document.querySelectorAll('input[type="radio"]').forEach(radio => radio.checked = false);
        document.querySelectorAll('input[type="checkbox"]').forEach(checkbox => checkbox.checked = false);
        document.querySelectorAll('.color-option').forEach(color => color.classList.remove('selected'));
        document.getElementById('minPrice').value = '';
        document.getElementById('maxPrice').value = '';
        document.getElementById('inStockOnly').checked = false;

        this.currentPage = 1;
        this.loadProducts();
    }

    toggleViewMode() {
        this.viewMode = this.viewMode === 'grid' ? 'list' : 'grid';
        const grid = document.getElementById('productsGrid');
        const icon = document.getElementById('viewIcon');
        
        grid.className = `products-grid ${this.viewMode}-view`;
        icon.className = this.viewMode === 'grid' ? 'fas fa-th' : 'fas fa-list';
    }

    toggleMobileFilters() {
        document.getElementById('filtersSidebar').classList.toggle('active');
    }

    closeMobileFilters() {
        document.getElementById('filtersSidebar').classList.remove('active');
    }

    goToPage(page) {
        this.currentPage = page;
        this.loadProducts();
        window.scrollTo({ top: 0, behavior: 'smooth' });
    }

    previousPage() {
        if (this.currentPage > 1) {
            this.goToPage(this.currentPage - 1);
        }
    }

    nextPage() {
        // We don't know total pages here, but the button will be disabled if needed
        this.goToPage(this.currentPage + 1);
    }

    updateProductsCount(total) {
        document.getElementById('productsCount').textContent = total;
    }

    updateCartCount() {
        // This will be handled by main.js
        if (window.app) {
            window.app.loadCartCount();
        }
    }

    updateNavbar() {
        // This will be handled by main.js
        if (window.app) {
            window.app.updateNavbar();
        }
    }

    formatPrice(price) {
        return new Intl.NumberFormat('ru-RU').format(price);
    }
}

// Initialize products manager when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.products = new ProductsManager();
});