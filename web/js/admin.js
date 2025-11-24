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
        this.initColors(); 
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
// Load available colors
    async loadColors() {
        try {
            const container = document.getElementById('colorsContainer');
            container.innerHTML = '<div class="loading-colors">Загрузка цветов...</div>';

            const colors = await api.request('/colors');
            
            if (!colors || colors.length === 0) {
                container.innerHTML = '<div class="no-colors">Цвета не найдены</div>';
                return;
            }

            this.allColors = colors;
            this.renderColors(colors);

        } catch (error) {
            console.error('Failed to load colors:', error);
            const container = document.getElementById('colorsContainer');
            container.innerHTML = '<div class="error">Не удалось загрузить цвета</div>';
        }
    }
     // Render available colors
    renderColors(colors) {
        const container = document.getElementById('colorsContainer');
        
        container.innerHTML = colors.map(color => `
            <div class="color-option" 
                 style="background-color: ${color.value}" 
                 title="${color.name}"
                 data-color-id="${color.id}"
                 data-color-name="${color.name}"
                 data-color-value="${color.value}"
                 onclick="adminProducts.toggleColorSelection(this)">
            </div>
        `).join('');

        // Mark already selected colors
        this.selectedColors.forEach(selectedColor => {
            const colorOption = container.querySelector(`[data-color-id="${selectedColor.id}"]`);
            if (colorOption) {
                colorOption.classList.add('selected');
            }
        });
    }
    // Toggle color selection
    toggleColorSelection(element) {
        const colorId = parseInt(element.dataset.colorId);
        const colorName = element.dataset.colorName;
        const colorValue = element.dataset.colorValue;

        const isSelected = element.classList.contains('selected');
        
        if (isSelected) {
            // Remove color
            element.classList.remove('selected');
            this.selectedColors = this.selectedColors.filter(color => color.id !== colorId);
        } else {
            // Add color
            element.classList.add('selected');
            this.selectedColors.push({
                id: colorId,
                name: colorName,
                value: colorValue
            });
        }

        this.renderSelectedColors();
    }
    // Render selected colors list
    renderSelectedColors() {
        const container = document.getElementById('selectedColorsList');
        
        if (this.selectedColors.length === 0) {
            container.innerHTML = '<div class="no-colors">Цвета не выбраны</div>';
            return;
        }

        container.innerHTML = this.selectedColors.map(color => `
            <div class="selected-color-item">
                <div class="selected-color-preview" style="background-color: ${color.value}"></div>
                <span class="selected-color-name">${color.name}</span>
                <button type="button" class="remove-color" onclick="adminProducts.removeSelectedColor(${color.id})">
                    <i class="fas fa-times"></i>
                </button>
            </div>
        `).join('');
    }
    // Remove selected color
    removeSelectedColor(colorId) {
        this.selectedColors = this.selectedColors.filter(color => color.id !== colorId);
        
        // Update visual selection
        const colorOption = document.querySelector(`[data-color-id="${colorId}"]`);
        if (colorOption) {
            colorOption.classList.remove('selected');
        }
        
        this.renderSelectedColors();
    }

    // Show add color modal
    showAddColorForm() {
        document.getElementById('addColorModal').style.display = 'block';
        document.getElementById('colorName').value = '';
        document.getElementById('colorValue').value = '#FF0000';
        document.getElementById('colorPicker').value = '#FF0000';
        document.getElementById('colorActive').checked = true;
        this.updateColorPreview();
    }

    // Close add color modal
    closeAddColorModal() {
        document.getElementById('addColorModal').style.display = 'none';
        this.clearColorFormErrors();
    }

    // Update color value from picker
    updateColorValue() {
        const picker = document.getElementById('colorPicker');
        const valueInput = document.getElementById('colorValue');
        valueInput.value = picker.value.toUpperCase();
        this.updateColorPreview();
    }
// Update color preview
    updateColorPreview() {
        const colorValue = document.getElementById('colorValue').value;
        const colorName = document.getElementById('colorName').value || 'Новый цвет';
        
        const previewBox = document.getElementById('colorPreviewBox');
        const previewText = document.getElementById('colorPreviewText');
        
        previewBox.style.backgroundColor = colorValue;
        previewText.textContent = colorName;
    }

    // Clear color form errors
    clearColorFormErrors() {
        document.getElementById('colorNameError').textContent = '';
        document.getElementById('colorValueError').textContent = '';
    }
    // Handle color form submission
    async handleAddColorForm(event) {
        event.preventDefault();
        
        const formData = {
            name: document.getElementById('colorName').value.trim(),
            value: document.getElementById('colorValue').value.toUpperCase(),
            active: document.getElementById('colorActive').checked
        };
    
    // Validation
        if (!formData.name) {
            document.getElementById('colorNameError').textContent = 'Введите название цвета';
            return;
        }

        if (!formData.value.match(/^#[0-9A-F]{6}$/i)) {
            document.getElementById('colorValueError').textContent = 'Введите корректный HEX код (например: #FF0000)';
            return;
        }

        try {
            const newColor = await api.request('/colors', {
                method: 'POST',
                body: {
                    name: formData.name,
                    value: formData.value,
                    active: formData.active
                }
            });

            // Add new color to available colors
            this.allColors.push(newColor);
            this.renderColors(this.allColors);
            
            // Select the new color automatically
            this.selectedColors.push({
                id: newColor.id,
                name: newColor.name,
                value: newColor.value
            });
            this.renderSelectedColors();
            
            // Close modal and show success message
            this.closeAddColorModal();
            this.showNotification('Цвет успешно добавлен', 'success');
            
        } catch (error) {
            console.error('Failed to create color:', error);
            this.showNotification('Ошибка при добавлении цвета: ${error.message}', 'error');
        }
    }

    // Get selected colors for form submission
    getSelectedColors() {
        return this.selectedColors.map(color => color.id);
    }
    // Set colors when editing product
    setProductColors(colorIds) {
        this.selectedColors = [];
        
        colorIds.forEach(colorId => {
            const color = this.allColors.find(c => c.id === colorId);
            if (color) {
                this.selectedColors.push({
                    id: color.id,
                    name: color.name,
                    value: color.value
                });
            }
        });
        
        this.renderSelectedColors();
        
        // Update visual selection
        this.allColors.forEach(color => {
            const colorOption = document.querySelector(`[data-color-id="${color.id}"]`);
            if (colorOption) {
                if (colorIds.includes(color.id)) {
                    colorOption.classList.add('selected');
                } else {
                    colorOption.classList.remove('selected');
                }
            }
        });
    }
    initColors() {
        this.loadColors();
        
        // Add event listener for color form
        document.getElementById('addColorForm').addEventListener('submit', (e) => {
            this.handleAddColorForm(e);
        });
        
        // Add event listeners for real-time preview
        document.getElementById('colorName').addEventListener('input', () => {
            this.updateColorPreview();
        });
        
        document.getElementById('colorValue').addEventListener('input', () => {
            this.updateColorPreview();
        });
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
            this.categories = data || [];

            this.renderCategoryOptions();
        } catch (err) {
            console.warn('Не удалось загрузить категории');
        }
    }

    renderCategoryOptions() {
        const el = document.getElementById('productCategory');
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
    async handleFormSubmit(e) {
        e.preventDefault();

        if (!this.validateForm()) {
            return;
        }

        const formData = this.getFormData();
        const submitBtn = document.getElementById('submitBtn');
        const originalText = submitBtn.innerHTML;

        try {
            submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Сохранение...';
            submitBtn.disabled = true;

            let response;
            if (this.currentProductId) {
                response = await this.apiRequest(`/admin/products/${this.currentProductId}`, 'PUT', formData);
            } else {
                response = await this.apiRequest('/admin/products', 'POST', formData);
            }

            this.showNotification(
                this.currentProductId ? 'Товар успешно обновлен' : 'Товар успешно создан',
                'success'
            );

            this.closeModal();
            this.loadProducts();

        } catch (error) {
            this.showNotification(error.message, 'error');
            this.displayFormErrors(error.errors || {});
        } finally {
            submitBtn.innerHTML = originalText;
            submitBtn.disabled = false;
        }
    }

    getFormData() {
        return {
            name: document.getElementById('productName').value,
            slug: document.getElementById('productSlug').value,
            description: document.getElementById('productDescription').value,
            sku: document.getElementById('productSKU').value,
            category_id: parseInt(document.getElementById('productCategory').value),
            price: parseFloat(document.getElementById('productPrice').value),
            compare_price: document.getElementById('productComparePrice').value ?
                parseFloat(document.getElementById('productComparePrice').value) : 0,
            in_stock: document.getElementById('productInStock').checked,
            is_active: document.getElementById('productIsActive').checked,
            images: this.uploadedImages,
            features: this.features,
            variations: this.variations
        };
    }

    validateForm() {
        let isValid = true;
        this.clearErrors();

        const requiredFields = [
            'productName', 'productSlug', 'productSKU', 'productCategory', 'productPrice'
        ];

        requiredFields.forEach(fieldId => {
            const field = document.getElementById(fieldId);
            if (!field.value.trim()) {
                this.showFieldError(fieldId, 'Это поле обязательно для заполнения');
                isValid = false;
            }
        });

        // Validate slug format
        const slug = document.getElementById('productSlug').value;
        if (slug && !/^[a-z0-9-]+$/.test(slug)) {
            this.showFieldError('productSlug', 'Slug может содержать только латинские буквы, цифры и дефисы');
            isValid = false;
        }

        return isValid;
    }

    generateSlug(name) {
        const slug = name
            .toLowerCase()
            .replace(/[^a-z0-9а-яё\s-]/g, '')
            .replace(/\s+/g, '-')
            .replace(/-+/g, '-')
            .replace(/^-|-$/g, '');

        document.getElementById('productSlug').value = slug;
    }

    // Image handling
    handleImageUpload(event) {
        const files = event.target.files;

        for (let file of files) {
            if (!file.type.startsWith('image/')) {
                this.showNotification('Пожалуйста, выбирайте только изображения', 'error');
                continue;
            }

            if (file.size > 5 * 1024 * 1024) { // 5MB
                this.showNotification('Размер изображения не должен превышать 5MB', 'error');
                continue;
            }

            this.previewImage(file);
        }

        event.target.value = ''; // Reset input
    }

    previewImage(file) {
        const reader = new FileReader();

        reader.onload = (e) => {
            const imageData = e.target.result;
            this.uploadedImages.push(imageData);
            this.renderImagePreviews();
        };

        reader.readAsDataURL(file);
    }

    renderImagePreviews() {
        const container = document.getElementById('imagePreview');

        container.innerHTML = this.uploadedImages.map((image, index) => `
            <div class="image-preview-item">
                <img src="${image}" alt="Preview ${index + 1}">
                <button type="button" class="remove-image" onclick="adminProducts.removeImage(${index})">
                    <i class="fas fa-times"></i>
                </button>
            </div>
        `).join('');
    }

    removeImage(index) {
        this.uploadedImages.splice(index, 1);
        this.renderImagePreviews();
    }

    // Features handling
    addFeature() {
        const input = document.getElementById('newFeature');
        const feature = input.value.trim();

        if (feature && !this.features.includes(feature)) {
            this.features.push(feature);
            this.renderFeatures();
            input.value = '';
        }
    }

    removeFeature(index) {
        this.features.splice(index, 1);
        this.renderFeatures();
    }

    renderFeatures() {
        const container = document.getElementById('featuresList');

        container.innerHTML = this.features.map((feature, index) => `
            <div class="feature-tag">
                ${feature}
                <button type="button" class="remove-feature" onclick="adminProducts.removeFeature(${index})">
                    <i class="fas fa-times"></i>
                </button>
            </div>
        `).join('');
    }

    // Variations handling
    addVariation() {
        this.variations.push({
            size_id: '',
            color_id: '',
            quantity: 0,
            image_url: ''
        });
        this.renderVariations();
    }

    removeVariation(index) {
        this.variations.splice(index, 1);
        this.renderVariations();
    }

    updateVariation(index, field, value) {
        this.variations[index][field] = value;
    }

    renderVariations() {
        const container = document.getElementById('variationsList');

        container.innerHTML = this.variations.map((variation, index) => `
            <div class="variation-item">
                <div class="variation-fields">
                    <div class="form-group">
                        <label>Размер</label>
                        <select onchange="adminProducts.updateVariation(${index}, 'size_id', this.value)">
                            <option value="">Выберите размер</option>
                            ${this.sizes ? this.sizes.map(size => `
                                <option value="${size.id}" ${variation.size_id === size.id ? 'selected' : ''}>
                                    ${size.name}
                                </option>
                            `).join('') : ''}
                        </select>
                    </div>
                    
                    <div class="form-group">
                        <label>Цвет</label>
                        <select onchange="adminProducts.updateVariation(${index}, 'color_id', this.value)">
                            <option value="">Выберите цвет</option>
                            ${this.colors ? this.colors.map(color => `
                                <option value="${color.id}" ${variation.color_id === color.id ? 'selected' : ''}>
                                    ${color.name}
                                </option>
                            `).join('') : ''}
                        </select>
                    </div>
                    
                    <div class="form-group">
                        <label>Количество</label>
                        <input type="number" value="${variation.quantity}" 
                               onchange="adminProducts.updateVariation(${index}, 'quantity', this.value)">
                    </div>
                    
                    <button type="button" class="remove-variation" onclick="adminProducts.removeVariation(${index})">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
            </div>
        `).join('');
    }

    async handleSubmit(e) {
        e.preventDefault();

        // Валидация формы
        if (!this.validateForm()) {
            return;
        }

        const form = e.target;
        const submitBtn = form.querySelector('button[type="submit"]');
        const originalText = submitBtn.innerHTML;
        const isEditMode = form.dataset.productId;

        try {
            // Показываем состояние загрузки
            this.setLoading(submitBtn, true);

            // Подготавливаем данные формы
            const formData = this.prepareFormData();

            let response;
            if (isEditMode) {
                // Редактирование существующего товара
                response = await api.put(`/admin/products/${form.dataset.productId}`, formData);
            } else {
                // Создание нового товара
                response = await api.post('/admin/products', formData);
            }

            // Успешное сохранение
            this.showNotification(
                isEditMode ? 'Товар успешно обновлен' : 'Товар успешно создан',
                'success'
            );

            // Закрываем модальное окно
            this.closeModal();

            // Обновляем список товаров
            this.loadProducts();

        } catch (error) {
            console.error('Save product error:', error);

            // Обработка ошибок валидации
            if (error.errors) {
                this.displayFormErrors(error.errors);
            } else {
                this.showNotification(
                    error.message || 'Произошла ошибка при сохранении товара',
                    'error'
                );
            }
        } finally {
            // Восстанавливаем кнопку
            this.setLoading(submitBtn, false);
        }
    }
    // 🔧 Вспомогательные методы для handleSubmit

    prepareFormData() {
        const formData = {
            name: document.getElementById('productName').value.trim(),
            slug: document.getElementById('productSlug').value.trim(),
            description: document.getElementById('productDescription').value.trim(),
            sku: document.getElementById('productSKU').value.trim(),
            price: parseFloat(document.getElementById('productPrice').value) || 0,
            compare_price: document.getElementById('productComparePrice').value ?
                parseFloat(document.getElementById('productComparePrice').value) : null,
            category_id: document.getElementById('productCategory').value ?
                parseInt(document.getElementById('productCategory').value) : null,
            in_stock: document.getElementById('productInStock').checked,
            is_active: document.getElementById('productIsActive').checked,
            features: this.features || [],
            variations: this.variations || []
        };

        // Очистка от null/undefined значений
        Object.keys(formData).forEach(key => {
            if (formData[key] === null || formData[key] === undefined || formData[key] === '') {
                delete formData[key];
            }
        });

        return formData;
    }
    validateForm() {
        let isValid = true;
        this.clearErrors();

        const requiredFields = [
            { id: 'productName', name: 'Название товара' },
            { id: 'productSlug', name: 'Slug' },
            { id: 'productSKU', name: 'Артикул' },
            { id: 'productPrice', name: 'Цена' }
        ];

        // Проверка обязательных полей
        requiredFields.forEach(field => {
            const element = document.getElementById(field.id);
            if (!element.value.trim()) {
                this.showFieldError(field.id, `Поле "${field.name}" обязательно для заполнения`);
                isValid = false;
            }
        });

        // Валидация slug
        const slug = document.getElementById('productSlug').value;
        if (slug && !/^[a-z0-9-]+$/.test(slug)) {
            this.showFieldError('productSlug', 'Slug может содержать только латинские буквы в нижнем регистре, цифры и дефисы');
            isValid = false;
        }

        // Валидация цены
        const price = document.getElementById('productPrice').value;
        if (price && (isNaN(price) || parseFloat(price) < 0)) {
            this.showFieldError('productPrice', 'Цена должна быть положительным числом');
            isValid = false;
        }

        // Валидация категории
        const category = document.getElementById('productCategory').value;
        if (!category) {
            this.showFieldError('productCategory', 'Необходимо выбрать категорию');
            isValid = false;
        }

        return isValid;
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
    showFieldError(fieldId, message) {
        const field = document.getElementById(fieldId);
        const errorElement = document.getElementById(`${fieldId}Error`) || this.createErrorElement(fieldId);

        field.classList.add('error');
        errorElement.textContent = message;
        errorElement.style.display = 'block';
    }

    createErrorElement(fieldId) {
        const errorElement = document.createElement('div');
        errorElement.id = `${fieldId}Error`;
        errorElement.className = 'field-error';

        const field = document.getElementById(fieldId);
        field.parentNode.insertBefore(errorElement, field.nextSibling);

        return errorElement;
    }

    clearErrors() {
        // Убираем класс error со всех полей
        document.querySelectorAll('.error').forEach(el => {
            el.classList.remove('error');
        });

        // Скрываем все сообщения об ошибках
        document.querySelectorAll('.field-error').forEach(el => {
            el.style.display = 'none';
        });
    }
    setLoading(button, isLoading) {
        if (isLoading) {
            button.disabled = true;
            button.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Сохранение...';
            button.classList.add('loading');
        } else {
            button.disabled = false;
            button.innerHTML = button.dataset.originalText || 'Сохранить';
            button.classList.remove('loading');
        }
    }

    showNotification(message, type = 'info') {
        // Создаем или находим контейнер для уведомлений
        let container = document.getElementById('notifications');
        if (!container) {
            container = document.createElement('div');
            container.id = 'notifications';
            container.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            z-index: 10000;
        `;
            document.body.appendChild(container);
        }

        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.innerHTML = `
        <div class="notification-content">
            <i class="fas fa-${this.getNotificationIcon(type)}"></i>
            <span>${message}</span>
            <button onclick="this.parentElement.parentElement.remove()">
                <i class="fas fa-times"></i>
            </button>
        </div>
    `;

        container.appendChild(notification);

        // Автоматическое удаление через 5 секунд
        setTimeout(() => {
            if (notification.parentElement) {
                notification.remove();
            }
        }, 5000);
    }
    getNotificationIcon(type) {
        const icons = {
            'success': 'check-circle',
            'error': 'exclamation-circle',
            'warning': 'exclamation-triangle',
            'info': 'info-circle'
        };
        return icons[type] || 'info-circle';
    }

    closeModal() {
        const modal = document.getElementById('productFormModal');
        if (modal) {
            modal.style.display = 'none';
        }

        // Очищаем форму и состояния
        const form = document.getElementById('productForm');
        if (form) {
            form.reset();
            form.removeAttribute('data-product-id');
        }

        this.clearErrors();
        this.uploadedImages = [];
        this.features = [];
        this.variations = [];

        // Очищаем превью изображений
        const preview = document.getElementById('imagePreview');
        if (preview) {
            preview.innerHTML = '';
        }
    }
    displayFormErrors(errors) {
        this.clearErrors();

        Object.keys(errors).forEach(field => {
            const fieldId = this.mapFieldNameToId(field);
            this.showFieldError(fieldId, errors[field].join(', '));
        });
    }
}

// Глобальный экземпляр
window.adminProducts = new AdminProducts();
