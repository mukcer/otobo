class AdminProducts {
    constructor() {
        this.currentPage = 1;
        this.totalPages = 1;
        this.products = [];
        this.categories = [];

        this.defaultFilters = {
            category: '',
            status: [],
            minPrice: '',
            maxPrice: '',
            inStockOnly: false,
            sortBy: 'created_at',
            sortOrder: 'desc',
            search: ''
        };

        this.currentFilters = { ...this.defaultFilters };

        this.init();
    }

    init() {
        this.loadInitialData();
        this.fetchCategories();
        this.setupEventListeners();
        this.loadProducts();
    }

    loadInitialData() {
        const dataEl = document.getElementById('initialData');
        if (dataEl) {
            const data = JSON.parse(dataEl.textContent);
            this.currentPage = data.currentPage || 1;
            this.currentFilters.search = data.searchQuery || '';
            document.getElementById('searchProducts').value = this.currentFilters.search;
        }
    }

    setupEventListeners() {
        // Применить фильтры
        document.getElementById('applyFilters').addEventListener('click', () => this.applyFilters());
        document.getElementById('clearFilters').addEventListener('click', () => this.clearFilters());

        // Сортировка
        document.getElementById('sortBy').addEventListener('change', () => {
            this.currentFilters.sortBy = document.getElementById('sortBy').value;
            this.loadProducts();
        });

        document.getElementById('sortOrder').addEventListener('change', () => {
            this.currentFilters.sortOrder = document.getElementById('sortOrder').value;
            this.loadProducts();
        });

        // Поиск
        // Уже есть onkeyup в HTML

        // Мобильные фильтры
        const mobileToggle = document.getElementById('mobileFiltersToggle');
        const sidebar = document.getElementById('filtersSidebar');
        if (mobileToggle && sidebar) {
            mobileToggle.addEventListener('click', () => {
                sidebar.classList.toggle('active');
            });
        }

        // Закрытие по клику вне
        document.addEventListener('click', (e) => {
            if (!e.target.closest('.filters-sidebar') && !e.target.closest('#mobileFiltersToggle')) {
                sidebar.classList.remove('active');
            }
        });

        // Закрытие модальных окон
        document.querySelectorAll('.close-modal').forEach(btn => {
            btn.addEventListener('click', () => {
                btn.closest('.modal').style.display = 'none';
            });
        });

        // Форма товара
        const form = document.getElementById('productForm');
        if (form) {
            form.addEventListener('submit', (e) => this.handleSubmit(e));
        }

        // Удаление товара
        document.getElementById('confirmDeleteBtn').addEventListener('click', () => {
            this.performDelete();
        });
    }

    applyFilters() {
        const categorySelect = document.getElementById('categoryFilter');
        if (categorySelect) {
            this.currentFilters.category = categorySelect.value;
        }

        this.currentFilters.status = [];
        if (document.getElementById('filterActive').checked) this.currentFilters.status.push('active');
        if (document.getElementById('filterInactive').checked) this.currentFilters.status.push('inactive');
        if (document.getElementById('filterOutOfStock').checked) this.currentFilters.status.push('out_of_stock');

        this.currentFilters.minPrice = document.getElementById('minPrice').value;
        this.currentFilters.maxPrice = document.getElementById('maxPrice').value;
        this.currentFilters.inStockOnly = document.getElementById('inStockOnly').checked;

        this.loadProducts();
    }

    clearFilters() {
        document.getElementById('categoryFilter').value = '';
        document.getElementById('minPrice').value = '';
        document.getElementById('maxPrice').value = '';
        document.getElementById('filterActive').checked = false;
        document.getElementById('filterInactive').checked = false;
        document.getElementById('filterOutOfStock').checked = false;
        document.getElementById('inStockOnly').checked = false;

        this.currentFilters = { ...this.defaultFilters };
        document.getElementById('searchProducts').value = '';
        this.loadProducts();
    }

    searchProducts(event) {
        clearTimeout(this.searchTimeout);
        this.searchTimeout = setTimeout(() => {
            this.currentFilters.search = event.target.value.trim();
            this.loadProducts();
        }, 500);
    }

    async loadProducts(page = 1) {
        this.currentPage = page;

        const url = new URL('/api/v1/admin/products', window.location.origin);
        const params = {
            page: this.currentPage,
            limit: 10,
            search: this.currentFilters.search,
            category_id: this.currentFilters.category,
            min_price: this.currentFilters.minPrice,
            max_price: this.currentFilters.maxPrice,
            in_stock: this.currentFilters.inStockOnly ? 1 : '',
            sort_by: this.currentFilters.sortBy,
            sort_order: this.currentFilters.sortOrder
        };

        if (this.currentFilters.status.length) {
            params.status = this.currentFilters.status.join(',');
        }

        Object.keys(params).forEach(key => {
            if (params[key]) {
                url.searchParams.append(key, params[key]);
            }
        });

        try {
            const data = await api.getProducts(params);

            this.renderProducts(data.products);
            this.renderPagination(data.pages);
            this.updateStats(data.products);
            } catch (err) {
            this.showError('Не удалось загрузить товары');
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

        // tbody.innerHTML = this.products.map(p => `
        //     <tr data-id="${p.id}">
        //         <td>${p.id}</td>
        //         <td>
        //             <div class="product-image">
        //                 <img src="${p.image || '/static/images/placeholder.jpg'}" alt="${p.name}" onerror="this.src='/static/images/placeholder.jpg'">
        //             </div>
        //         </td>
        //         <td>
        //             <div class="product-info">
        //                 <strong>${this.escapeHtml(p.name)}</strong>
        //                 <small>Артикул: ${p.sku}</small>
        //             </div>
        //         </td>
        //         <td>${this.escapeHtml(p.category?.name || '–')}</td>
        //         <td>${p.price} ₽</td>
        //         <td>
        //             <span class="status-badge ${p.is_active ? 'active' : 'inactive'}">
        //                 ${p.is_active ? 'Активен' : 'Неактивен'}
        //             </span>
        //         </td>
        //         <td>${p.stock || 0}</td>
        //         <td>${this.formatDate(p.created_at)}</td>
        //         <td class="actions">
        //             <button class="action-btn view" title="Просмотр" onclick="adminProducts.openViewModal(${p.id})">
        //                 <i class="fas fa-eye"></i>
        //             </button>
        //             <button class="action-btn edit" title="Редактировать" onclick="adminProducts.openEditModal(${p.id})">
        //                 <i class="fas fa-edit"></i>
        //             </button>
        //             <button class="action-btn delete" title="Удалить" onclick="adminProducts.openDeleteModal(${p.id})">
        //                 <i class="fas fa-trash"></i>
        //             </button>
        //         </td>
        //     </tr>
        // `).join('');
        grid.innerHTML = products.map(product => `
            <div class="product-card" onclick="products.openProductModal('${product.id}')">
                <div class="product-image">
                    ${product.images && product.images.length > 0 ? 
                        `<img src="${product.images[0] || '/static/images/placeholder.jpg'}" alt="${product.name}" loading="lazy">` :
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
                <div class="actions">
                    <button class="action-btn view" title="Просмотр" onclick="adminProducts.openViewModal(${product.id})">
                         <i class="fas fa-eye"></i>
                     </button>
                     <button class="action-btn edit" title="Редактировать" onclick="adminProducts.openEditModal(${product.id})">
                         <i class="fas fa-edit"></i>
                     </button>
                     <button class="action-btn delete" title="Удалить" onclick="adminProducts.openDeleteModal(${product.id})">
                         <i class="fas fa-trash"></i>
                     </button>
                </div>
            </div>
        `).join('');
        // Apply view mode
        grid.className = `products-grid ${this.viewMode}-view`;
    }

    renderPagination(totalPages) {
        const container = document.getElementById('pagination');
        container.innerHTML = '';

        if (totalPages <= 1) return;

        const fragment = document.createDocumentFragment();

        // Prev
        if (this.currentPage > 1) {
            const prev = this.createPageButton('«', this.currentPage - 1);
            fragment.appendChild(prev);
        }

        // First
        if (this.currentPage > 2) {
            const first = this.createPageButton('1', 1);
            fragment.appendChild(first);
            if (this.currentPage > 3) {
                const ellipsis = document.createElement('span');
                ellipsis.textContent = '...';
                fragment.appendChild(ellipsis);
            }
        }

        // Current range
        const start = Math.max(1, this.currentPage - 1);
        const end = Math.min(this.totalPages, this.currentPage + 1);

        for (let i = start; i <= end; i++) {
            const btn = this.createPageButton(i, i, i === this.currentPage);
            fragment.appendChild(btn);
        }

        // Last
        if (this.currentPage < this.totalPages - 1) {
            if (this.currentPage < this.totalPages - 2) {
                const ellipsis = document.createElement('span');
                ellipsis.textContent = '...';
                fragment.appendChild(ellipsis);
            }
            const last = this.createPageButton(this.totalPages, this.totalPages);
            fragment.appendChild(last);
        }

        // Next
        if (this.currentPage < this.totalPages) {
            const next = this.createPageButton('»', this.currentPage + 1);
            fragment.appendChild(next);
        }

        container.appendChild(fragment);
    }

    createPageButton(label, page, active = false) {
        const btn = document.createElement('button');
        btn.textContent = label;
        btn.className = active ? 'active' : '';
        btn.onclick = () => this.loadProducts(page);
        return btn;
    }

    updateStats(products) {
        const total = products.length;
        const active = products.filter(p => p.is_active).length;
        const outOfStock = products.filter(p => p.stock <= 0).length;

        document.getElementById('totalProducts').textContent = total;
        document.getElementById('activeProducts').textContent = active;
        document.getElementById('outOfStock').textContent = outOfStock;
    }

    async fetchCategories() {
        try {
            const data = await api.getCategories();
            this.categories = data.data || [];

            this.renderCategoryOptions();
        } catch (err) {
            console.warn('Не удалось загрузить категории');
        }
    }

    renderCategoryOptions() {
        const el = document.getElementById('categoryFilter');
        if (!el) return;

        el.innerHTML = '<option value="">Все категории</option>';
        this.categories.forEach(cat => {
            const option = document.createElement('option');
            option.value = cat.id;
            option.textContent = cat.name;
            el.appendChild(option);
        });
    }

    async openViewModal(id) {
        const product = await api.getProductByID(id);
        if (!product) return;

        const modalBody = document.getElementById('modalBody');
        modalBody.innerHTML = `
            <div class="product-detail">
                <div class="detail-image">
                    <img src="${product.image || '/static/images/placeholder.jpg'}" alt="${product.name}">
                </div>
                <div class="detail-info">
                    <h3>${this.escapeHtml(product.name)}</h3>
                    <p class="detail-description">${this.escapeHtml(product.description || '–')}</p>
                    
                    <div class="detail-meta">
                        <div><strong>Цена:</strong> ${product.price} ₽</div>
                        <div><strong>Артикул:</strong> ${product.sku}</div>
                        <div><strong>Категория:</strong> ${this.escapeHtml(product.category?.name || '–')}</div>
                        <div><strong>Наличие:</strong> ${product.stock > 0 ? `${product.stock} шт.` : '<span class="out-of-stock">Нет в наличии</span>'}</div>
                        <div><strong>Дата:</strong> ${this.formatDate(product.created_at)}</div>
                    </div>
                </div>
            </div>
        `;

        document.getElementById('productModal').style.display = 'block';
    }

    async openEditModal(id) {
        this.openCreateModal();
        document.getElementById('modalTitle').textContent = 'Редактировать товар';
        document.getElementById('submitBtn').textContent = 'Сохранить изменения';

        try {
            const product = await api.getProductByID(id);

        // Заполняем поля формы
        document.getElementById('productName').value = product.name || '';
        document.getElementById('productSlug').value = product.slug || '';
        document.getElementById('productSKU').value = product.sku || '';
        document.getElementById('productDescription').value = product.description || '';
        document.getElementById('productPrice').value = product.price || '';
        document.getElementById('productComparePrice').value = product.compare_price || '';

        // Установка категории (если есть <select id="productCategory">)
        const categorySelect = document.getElementById('productCategory');
        if (categorySelect) {
            for (let option of categorySelect.options) {
                if (option.value == product.category_id) {
                    option.selected = true;
                    break;
                }
            }
        }

        // Чекбоксы
        document.getElementById('productInStock').checked = !!product.in_stock;
        document.getElementById('productIsActive').checked = !!product.is_active;

        // Очистка предыдущих изображений
        document.getElementById('imagePreview').innerHTML = '';

        // Предпросмотр изображений
        if (product.images && product.images.length > 0) {
            product.images.forEach(img => {
                const preview = document.createElement('div');
                preview.className = 'image-item';
                preview.innerHTML = `
                    <img src="${img.url}" alt="image" style="width: 100px; height: 100px; object-fit: cover;">
                    <button type="button" class="remove-image" onclick="this.parentElement.remove()">
                        <i class="fas fa-times"></i>
                    </button>
                `;
                document.getElementById('imagePreview').appendChild(preview);
            });
        }

        // Сохраняем ID товара для последующего PUT-запроса
        document.getElementById('productForm').dataset.productId = product.id;

    } catch (error) {
        console.error('Failed to load product:', error);
        this.showError('Не удалось загрузить товар для редактирования');
    }
}

    openCreateModal() {
        document.getElementById('modalTitle').textContent = 'Добавить товар';
        document.getElementById('submitBtn').textContent = 'Создать товар';
        document.getElementById('productForm').reset();
        document.getElementById('imagePreview').innerHTML = '';
        document.getElementById('productFormModal').style.display = 'block';
    }

    openDeleteModal(id) {
        document.getElementById('deleteModal').style.display = 'block';
        document.getElementById('confirmDeleteBtn').dataset.id = id;
    }

    closeDeleteModal() {
        document.getElementById('deleteModal').style.display = 'none';
    }

    async performDelete() {
        const id = document.getElementById('confirmDeleteBtn').dataset.id;
        if (!id) return;

        try {
            const response = await fetch(`/api/v1/admin/products/${id}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });

            if (response.ok) {
                this.closeDeleteModal();
                this.loadProducts();
            } else {
                this.showError('Не удалось удалить товар');
            }
        } catch (err) {
            this.showError('Ошибка соединения');
        }
    }

    async handleSubmit(e) {
        e.preventDefault();
        // Здесь будет логика сохранения
        alert('Форма сохранения временно недоступна');
    }

    showError(message) {
        alert(message);
        // Можно сделать красивый тост
    }

    formatDate(dateStr) {
        const date = new Date(dateStr);
        return date.toLocaleDateString('ru-RU');
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
    formatPrice(price) {
        return new Intl.NumberFormat('ru-RU').format(price);
    }
}

// Глобальный экземпляр
window.adminProducts = new AdminProducts();
