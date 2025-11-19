class AdminProductsManager {
    constructor() {
        this.currentPage = 1;
        this.limit = 10;
        this.filters = {
            search: '',
            category: '',
            status: ''
        };
        this.currentProductId = null;
        this.uploadedImages = [];
        this.features = [];
        this.variations = [];
        
        this.init();
    }

    init() {
        this.loadProducts();
        this.loadCategories();
        this.setupEventListeners();
    }

    setupEventListeners() {
        // Form submission
        document.getElementById('productForm').addEventListener('submit', (e) => this.handleFormSubmit(e));
        
        // Auto-generate slug from name
        document.getElementById('productName').addEventListener('input', (e) => {
            if (!this.currentProductId) { // Only auto-generate for new products
                this.generateSlug(e.target.value);
            }
        });
    }

    async loadProducts() {
        const tbody = document.getElementById('productsTableBody');
        tbody.innerHTML = `
            <tr>
                <td colspan="9" class="loading-row">
                    <i class="fas fa-spinner fa-spin"></i>
                    Загрузка товаров...
                </td>
            </tr>
        `;

        try {
            const params = {
                page: this.currentPage,
                limit: this.limit,
                ...this.filters
            };

            // Remove empty filters
            Object.keys(params).forEach(key => {
                if (!params[key]) delete params[key];
            });

            const response = await this.apiRequest('/admin/products', 'GET', null, params);
            this.renderProductsTable(response.products);
            this.renderPagination(response.total, response.pages);
            this.updateStats(response.stats);

        } catch (error) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="9" class="loading-row">
                        <i class="fas fa-exclamation-triangle"></i>
                        Ошибка загрузки: ${error.message}
                    </td>
                </tr>
            `;
        }
    }

    renderProductsTable(products) {
        const tbody = document.getElementById('productsTableBody');
        
        if (products.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="9" class="loading-row">
                        Товары не найдены
                    </td>
                </tr>
            `;
            return;
        }

        tbody.innerHTML = products.map(product => `
            <tr>
                <td>${product.id}</td>
                <td>
                    <div class="product-image-small">
                        ${product.images && product.images.length > 0 ? 
                            `<img src="${product.images[0]}" alt="${product.name}">` :
                            `<i class="fas fa-image"></i>`
                        }
                    </div>
                </td>
                <td>
                    <strong>${product.name}</strong>
                    <div class="product-sku">${product.sku}</div>
                </td>
                <td>${product.category?.name || '—'}</td>
                <td>${this.formatPrice(product.price)} ₽</td>
                <td>
                    <span class="status-badge ${this.getStatusClass(product)}">
                        ${this.getStatusText(product)}
                    </span>
                </td>
                <td>
                    ${this.getTotalStock(product.variations)}
                </td>
                <td>${this.formatDate(product.created_at)}</td>
                <td>
                    <div class="action-buttons">
                        <button class="btn-sm btn-edit" onclick="adminProducts.openEditModal(${product.id})">
                            <i class="fas fa-edit"></i>
                            Редакт.
                        </button>
                        <button class="btn-sm btn-delete" onclick="adminProducts.openDeleteModal(${product.id})">
                            <i class="fas fa-trash"></i>
                            Удалить
                        </button>
                    </div>
                </td>
            </tr>
        `).join('');
    }

    getStatusClass(product) {
        if (!product.is_active) return 'status-inactive';
        if (!product.in_stock) return 'status-out-of-stock';
        return 'status-active';
    }

    getStatusText(product) {
        if (!product.is_active) return 'Неактивен';
        if (!product.in_stock) return 'Нет в наличии';
        return 'Активен';
    }

    getTotalStock(variations) {
        if (!variations || variations.length === 0) return 0;
        return variations.reduce((sum, variation) => sum + (variation.quantity || 0), 0);
    }

    async openCreateModal() {
        this.currentProductId = null;
        document.getElementById('modalTitle').textContent = 'Добавить товар';
        document.getElementById('submitBtn').innerHTML = '<i class="fas fa-save"></i> Сохранить товар';
        
        this.resetForm();
        this.loadCategoriesForForm();
        await this.loadSizesAndColors();
        
        document.getElementById('productFormModal').style.display = 'block';
    }

    async openEditModal(productId) {
        try {
            const product = await this.apiRequest(`/admin/products/${productId}`);
            
            this.currentProductId = productId;
            document.getElementById('modalTitle').textContent = 'Редактировать товар';
            document.getElementById('submitBtn').innerHTML = '<i class="fas fa-save"></i> Обновить товар';
            
            this.populateForm(product);
            this.loadCategoriesForForm();
            await this.loadSizesAndColors();
            
            document.getElementById('productFormModal').style.display = 'block';
            
        } catch (error) {
            this.showNotification('Ошибка загрузки товара', 'error');
        }
    }

    populateForm(product) {
        document.getElementById('productName').value = product.name || '';
        document.getElementById('productSlug').value = product.slug || '';
        document.getElementById('productSKU').value = product.sku || '';
        document.getElementById('productDescription').value = product.description || '';
        document.getElementById('productPrice').value = product.price || '';
        document.getElementById('productComparePrice').value = product.compare_price || '';
        document.getElementById('productInStock').checked = product.in_stock || false;
        document.getElementById('productIsActive').checked = product.is_active !== false;

        // Set category
        setTimeout(() => {
            const categorySelect = document.getElementById('productCategory');
            if (product.category_id) {
                categorySelect.value = product.category_id;
            }
        }, 100);

        // Images
        this.uploadedImages = product.images || [];
        this.renderImagePreviews();

        // Features
        this.features = product.features || [];
        this.renderFeatures();

        // Variations
        this.variations = product.variations || [];
        this.renderVariations();
    }

    resetForm() {
        document.getElementById('productForm').reset();
        this.uploadedImages = [];
        this.features = [];
        this.variations = [];
        
        this.renderImagePreviews();
        this.renderFeatures();
        this.renderVariations();
        
        this.clearErrors();
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

    // API methods
    async apiRequest(endpoint, method = 'GET', data = null, params = null) {
        const url = new URL(`${window.location.origin}${endpoint}`);
        
        if (params) {
            Object.keys(params).forEach(key => {
                if (params[key]) url.searchParams.append(key, params[key]);
            });
        }

        const options = {
            method,
            headers: {
                'Content-Type': 'application/json',
            }
        };

        if (data && (method === 'POST' || method === 'PUT')) {
            options.body = JSON.stringify(data);
        }

        const response = await fetch(url, options);
        const result = await response.json();

        if (!response.ok) {
            throw new Error(result.error || 'Request failed');
        }

        return result;
    }

    // Utility methods
    showFieldError(fieldId, message) {
        const errorElement = document.getElementById(fieldId + 'Error');
        const fieldElement = document.getElementById(fieldId);
        
        if (errorElement && fieldElement) {
            errorElement.textContent = message;
            fieldElement.classList.add('error');
        }
    }

    clearErrors() {
        document.querySelectorAll('.error-message').forEach(el => el.textContent = '');
        document.querySelectorAll('.error').forEach(el => el.classList.remove('error'));
    }

    displayFormErrors(errors) {
        Object.keys(errors).forEach(field => {
            this.showFieldError('product' + this.capitalize(field), errors[field]);
        });
    }

    capitalize(str) {
        return str.charAt(0).toUpperCase() + str.slice(1);
    }

    showNotification(message, type = 'success') {
        // Use the notification system from main.js
        if (window.app && typeof window.app.showNotification === 'function') {
            window.app.showNotification(message, type);
        } else {
            alert(message);
        }
    }

    formatPrice(price) {
        return new Intl.NumberFormat('ru-RU').format(price);
    }

    formatDate(dateString) {
        return new Date(dateString).toLocaleDate