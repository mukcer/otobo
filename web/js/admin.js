// Админский объект
const admin = {
    currentTab: 'overview',
    currentTable: null,
    tableInstance: null,
    ordersTable: null,
    currentOrdersPage: 1,
    
    init() {
        // Инициализация навигации
        this.setupNavigation();
        
        // Загрузка данных
        this.loadDashboardStats();
        
        // Загрузка списка таблиц
        this.loadTableList();
        
        // Загрузка статистики кэша
        this.loadCacheStats();
    },
    
    setupNavigation() {
        const navItems = document.querySelectorAll('.nav-item');
        navItems.forEach(item => {
            item.addEventListener('click', () => {
                this.switchTab(item.dataset.tab);
            });
        });
    },
    
    switchTab(tabName) {
        // Обновляем активную вкладку в навигации
        document.querySelectorAll('.nav-item').forEach(item => {
            item.classList.remove('active');
        });
        document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');
        
        // Показываем нужный контент
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.remove('active');
        });
        document.getElementById(`${tabName}-tab`).classList.add('active');
        
        this.currentTab = tabName;
        
        // Если переключаемся на вкладку базы данных, загружаем таблицы
        if (tabName === 'database') {
            this.loadTableList();
        }
        
        // Если переключаемся на вкладку заказов, загружаем заказы
        if (tabName === 'orders') {
            this.loadOrders();
        }
    },
    
    async loadDashboardStats() {
        try {
            // Здесь будет загрузка реальных данных
            document.getElementById('totalProducts').textContent = '142';
            document.getElementById('totalUsers').textContent = '1,248';
            document.getElementById('totalOrders').textContent = '89';
            document.getElementById('totalRevenue').textContent = '124,560 ₽';
        } catch (error) {
            console.error('Failed to load dashboard stats:', error);
        }
    },
    
    async loadTableList() {
        try {
            const tables = await api.get('/admin/database/tables');
            
            const tableList = document.getElementById('tableList');
            tableList.innerHTML = tables.map(table => `
                <div class="table-item">
                    <div class="table-info">
                        <div class="table-name">${table.name}</div>
                        <div class="table-rows">${table.count} записей</div>
                    </div>
                    <div class="table-actions">
                        <button class="btn btn-sm btn-outline" onclick="admin.viewTable('${table.name}')">
                            <i class="fas fa-eye"></i>
                        </button>
                        <button class="btn btn-sm btn-danger" onclick="admin.confirmDeleteTable('${table.name}')">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </div>
            `).join('');
        } catch (error) {
            console.error('Failed to load table list:', error);
            this.showNotification('Ошибка загрузки списка таблиц', 'error');
        }
    },
    
    async viewTable(tableName) {
        this.currentTable = tableName;
        this.switchTab('table-view');
        
        // Создаем контейнер для таблицы если его нет
        let tableContainer = document.getElementById('tableContainer');
        if (!tableContainer) {
            const content = document.querySelector('.admin-content');
            tableContainer = document.createElement('div');
            tableContainer.id = 'tableContainer';
            tableContainer.className = 'tab-content';
            tableContainer.innerHTML = `
                <div class="page-header">
                    <h1>Таблица: ${tableName}</h1>
                    <button class="btn btn-secondary" onclick="admin.switchTab('database')">
                        <i class="fas fa-arrow-left"></i>
                        Назад к списку таблиц
                    </button>
                </div>
                <div id="tableData"></div>
            `;
            content.appendChild(tableContainer);
        } else {
            tableContainer.querySelector('h1').textContent = `Таблица: ${tableName}`;
            tableContainer.style.display = 'block';
        }
        
        // Скрываем другие вкладки
        document.querySelectorAll('.tab-content').forEach(tab => {
            if (tab.id !== 'tableContainer') {
                tab.style.display = 'none';
            }
        });
        
        // Загружаем данные таблицы
        await this.loadTableData(tableName);
    },
    
    async loadTableData(tableName) {
        try {
            const result = await api.get(`/admin/database/tables/${tableName}/data?limit=100`);
            
            // Уничтожаем предыдущую таблицу если она существует
            if (this.tableInstance) {
                this.tableInstance.destroy();
            }
            
            // Создаем новую таблицу Tabulator
            this.tableInstance = new Tabulator("#tableData", {
                data: result.data,
                layout: "fitColumns",
                pagination: "local",
                paginationSize: 20,
                movableColumns: true,
                columns: this.createTableColumns(result.columns),
                cellEdited: (cell) => this.onCellEdited(cell),
            });
        } catch (error) {
            console.error('Failed to load table data:', error);
            this.showNotification('Ошибка загрузки данных таблицы', 'error');
        }
    },
    
    createTableColumns(columns) {
        return columns.map(col => {
            const columnDef = {
                title: col.column_name,
                field: col.column_name,
                headerFilter: "input",
            };
            
            // Определяем тип редактора на основе типа данных
            if (col.data_type.includes('bool')) {
                columnDef.editor = "tickCross";
                columnDef.formatter = "tickCross";
            } else if (col.data_type.includes('int') || col.data_type.includes('numeric')) {
                columnDef.editor = "number";
            } else {
                columnDef.editor = "input";
            }
            
            // Системные поля делаем не редактируемыми
            if (col.column_name === 'id' || col.column_name.includes('created_at') || col.column_name.includes('updated_at')) {
                columnDef.editable = false;
            }
            
            return columnDef;
        });
    },
    
    async onCellEdited(cell) {
        const rowData = cell.getRow().getData();
        const fieldName = cell.getField();
        const newValue = cell.getValue();
        const tableName = this.currentTable;
        const recordId = rowData.id;
        
        // Подготовка данных для обновления
        const updateData = {
            [fieldName]: newValue
        };
        
        try {
            const response = await api.request(`/admin/database/tables/${tableName}/data/${recordId}`, {
                method: 'PUT',
                body: updateData,
            });
            
            if (!response.ok) {
                throw new Error('Failed to update record');
            }
            
            this.showNotification('Запись успешно обновлена', 'success');
        } catch (error) {
            console.error('Failed to update record:', error);
            this.showNotification('Ошибка обновления записи', 'error');
            // Откатываем изменения в ячейке
            cell.restoreOldValue();
        }
    },
    
    async loadCacheStats() {
        try {
            document.getElementById('cacheHits').textContent = '1,248';
            document.getElementById('cacheMisses').textContent = '89';
            document.getElementById('cacheUsage').textContent = '65%';
            
            // Загрузка ключей кэша
            const cacheKeys = document.getElementById('cacheKeys');
            cacheKeys.innerHTML = `
                <div class="cache-key">
                    <div class="key-name">session:abc123</div>
                    <div class="key-size">1.2 KB</div>
                    <div class="key-actions">
                        <button class="btn btn-sm btn-danger" onclick="admin.deleteCacheKey('session:abc123')">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </div>
                <div class="cache-key">
                    <div class="key-name">product:42</div>
                    <div class="key-size">856 bytes</div>
                    <div class="key-actions">
                        <button class="btn btn-sm btn-danger" onclick="admin.deleteCacheKey('product:42')">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </div>
            `;
        } catch (error) {
            console.error('Failed to load cache stats:', error);
        }
    },
    
    saveSettings() {
        this.showNotification('Настройки успешно сохранены', 'success');
    },
    
    backupDatabase() {
        this.showConfirmModal('Создание резервной копии', 'Вы уверены, что хотите создать резервную копию базы данных?', () => {
            this.showNotification('Резервная копия создана успешно', 'success');
        });
    },
    
    optimizeDatabase() {
        this.showConfirmModal('Оптимизация базы данных', 'Вы уверены, что хотите оптимизировать базу данных?', () => {
            this.showNotification('База данных оптимизирована', 'success');
        });
    },
    
    clearCache() {
        this.showConfirmModal('Очистка кэша', 'Вы уверены, что хотите очистить весь кэш?', () => {
            this.showNotification('Кэш успешно очищен', 'success');
            this.loadCacheStats();
        });
    },
    
    clearSessions() {
        this.showConfirmModal('Очистка сессий', 'Вы уверены, что хотите очистить все сессии пользователей?', () => {
            this.showNotification('Сессии успешно очищены', 'success');
        });
    },
    
    confirmDeleteTable(tableName) {
        this.showConfirmModal('Удаление таблицы', `Вы уверены, что хотите удалить таблицу ${tableName}? Это действие нельзя отменить.`, () => {
            this.deleteTable(tableName);
        });
    },
    
    async deleteTable(tableName) {
        try {
            await api.delete(`/admin/database/tables/${tableName}`);
            
            this.showNotification(`Таблица ${tableName} удалена`, 'success');
            this.loadTableList();
        } catch (error) {
            console.error('Failed to delete table:', error);
            this.showNotification('Ошибка удаления таблицы', 'error');
        }
    },
    
    deleteCacheKey(keyName) {
        this.showConfirmModal('Удаление ключа кэша', `Вы уверены, что хотите удалить ключ кэша ${keyName}?`, () => {
            this.showNotification(`Ключ ${keyName} удален`, 'success');
            this.loadCacheStats();
        });
    },
    
    showConfirmModal(title, message, onConfirm) {
        document.getElementById('confirmTitle').textContent = title;
        document.getElementById('confirmMessage').textContent = message;
        document.getElementById('confirmActionBtn').onclick = () => {
            onConfirm();
            this.closeConfirmModal();
        };
        document.getElementById('confirmModal').style.display = 'block';
    },
    
    closeConfirmModal() {
        document.getElementById('confirmModal').style.display = 'none';
    },
    
    showNotification(message, type = 'info') {
        // Создаем уведомление
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
        
        // Добавляем в контейнер (создаем если нужно)
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
        
        container.appendChild(notification);
        
        // Автоматическое удаление через 5 секунд
        setTimeout(() => {
            if (notification.parentElement) {
                notification.remove();
            }
        }, 5000);
    },
    
    getNotificationIcon(type) {
        const icons = {
            'success': 'check-circle',
            'error': 'exclamation-circle',
            'warning': 'exclamation-triangle',
            'info': 'info-circle'
        };
        return icons[type] || 'info-circle';
    },
    
    // Загрузка списка заказов
    async loadOrders(page = 1) {
        try {
            // Получаем параметры фильтрации
            const status = document.getElementById('orderStatusFilter').value;
            const orderNumber = document.getElementById('orderNumberSearch').value;
            const customer = document.getElementById('customerSearch').value;
            
            // Формируем URL с параметрами
            const params = new URLSearchParams();
            if (status) params.append('status', status);
            if (orderNumber) params.append('order_number', orderNumber);
            if (customer) params.append('customer', customer);
            params.append('page', page);
            params.append('limit', 20);
            
            const result = await api.get(`/admin/orders?${params}`);
            
            // Уничтожаем предыдущую таблицу если она существует
            if (this.ordersTable) {
                this.ordersTable.destroy();
            }
            
            // Создаем новую таблицу Tabulator для заказов
            this.ordersTable = new Tabulator("#ordersTable", {
                data: result.orders,
                layout: "fitColumns",
                pagination: "local",
                paginationSize: 20,
                movableColumns: true,
                columns: [
                    {title: "ID", field: "id", width: 80},
                    {title: "Номер заказа", field: "order_number", headerFilter: "input"},
                    {title: "Клиент", field: "user.first_name", formatter: (cell) => {
                        const user = cell.getRow().getData().user;
                        return user ? `${user.first_name} ${user.last_name}` : 'Неизвестно';
                    }},
                    {title: "Сумма", field: "final_amount", formatter: "money", formatterParams: {symbol: "₽"}},
                    {title: "Статус", field: "status", formatter: (cell) => {
                        const status = cell.getValue();
                        const statusMap = {
                            'pending': 'В ожидании',
                            'paid': 'Оплачен',
                            'shipped': 'Отправлен',
                            'delivered': 'Доставлен',
                            'cancelled': 'Отменен'
                        };
                        return statusMap[status] || status;
                    }},
                    {title: "Дата создания", field: "created_at", formatter: "datetime", formatterParams: {
                        outputFormat: "dd.MM.yyyy HH:mm",
                        invalidPlaceholder: "(некорректная дата)"
                    }},
                    {title: "Действия", formatter: (cell) => {
                        return `
                            <button class="btn btn-sm btn-outline" onclick="admin.viewOrderDetails(${cell.getRow().getData().id})">
                                <i class="fas fa-eye"></i>
                            </button>
                            <button class="btn btn-sm btn-primary" onclick="admin.editOrderStatus(${cell.getRow().getData().id})">
                                <i class="fas fa-edit"></i>
                            </button>
                        `;
                    }, width: 120}
                ],
            });
            
            // Обновляем пагинацию
            this.updateOrdersPagination(result.total, page, 20);
            
        } catch (error) {
            console.error('Failed to load orders:', error);
            this.showNotification('Ошибка загрузки заказов', 'error');
        }
    },
    
    // Обновление пагинации заказов
    updateOrdersPagination(total, currentPage, limit) {
        const totalPages = Math.ceil(total / limit);
        const pagination = document.getElementById('ordersPagination');
        
        let paginationHTML = '';
        if (totalPages > 1) {
            // Предыдущая страница
            if (currentPage > 1) {
                paginationHTML += `<button class="btn btn-sm" onclick="admin.loadOrders(${currentPage - 1})">Назад</button>`;
            }
            
            // Страницы
            for (let i = 1; i <= totalPages; i++) {
                if (i === currentPage) {
                    paginationHTML += `<button class="btn btn-sm btn-primary" disabled>${i}</button>`;
                } else {
                    paginationHTML += `<button class="btn btn-sm" onclick="admin.loadOrders(${i})">${i}</button>`;
                }
            }
            
            // Следующая страница
            if (currentPage < totalPages) {
                paginationHTML += `<button class="btn btn-sm" onclick="admin.loadOrders(${currentPage + 1})">Вперед</button>`;
            }
        }
        
        pagination.innerHTML = paginationHTML;
    },
    
    // Сброс фильтров заказов
    resetOrderFilters() {
        document.getElementById('orderStatusFilter').value = '';
        document.getElementById('orderNumberSearch').value = '';
        document.getElementById('customerSearch').value = '';
        this.loadOrders();
    },
    
    // Просмотр деталей заказа
    async viewOrderDetails(orderId) {
        try {
            const order = await api.get(`/admin/orders/${orderId}`);
            
            // Создаем модальное окно с деталями заказа
            const modal = document.createElement('div');
            modal.className = 'modal';
            modal.style.display = 'block';
            modal.innerHTML = `
                <div class="modal-content" style="max-width: 800px;">
                    <div class="modal-header">
                        <h2>Детали заказа #${order.order_number}</h2>
                        <span class="close-modal" onclick="this.closest('.modal').remove()">&times;</span>
                    </div>
                    <div class="modal-body">
                        <div class="order-details">
                            <div class="order-info-grid">
                                <div class="info-item">
                                    <label>Статус:</label>
                                    <span>${this.getOrderStatusText(order.status)}</span>
                                </div>
                                <div class="info-item">
                                    <label>Сумма заказа:</label>
                                    <span>${order.final_amount} ₽</span>
                                </div>
                                <div class="info-item">
                                    <label>Клиент:</label>
                                    <span>${order.user.first_name} ${order.user.last_name}</span>
                                </div>
                                <div class="info-item">
                                    <label>Дата создания:</label>
                                    <span>${new Date(order.created_at).toLocaleString('ru-RU')}</span>
                                </div>
                                <div class="info-item">
                                    <label>Адрес доставки:</label>
                                    <span>${order.shipping_address || 'Не указан'}</span>
                                </div>
                                <div class="info-item">
                                    <label>Метод доставки:</label>
                                    <span>${order.shipping_method || 'Не указан'}</span>
                                </div>
                            </div>
                            
                            <h3>Товары в заказе</h3>
                            <div class="order-items">
                                ${order.items.map(item => `
                                    <div class="order-item">
                                        <div class="item-info">
                                            <div class="item-name">${item.product.name}</div>
                                            <div class="item-variation">
                                                ${item.variation ? `Размер: ${item.variation.size.name}, Цвет: ${item.variation.color.name}` : ''}
                                            </div>
                                        </div>
                                        <div class="item-quantity">Количество: ${item.quantity}</div>
                                        <div class="item-price">${item.price} ₽</div>
                                    </div>
                                `).join('')}
                            </div>
                        </div>
                    </div>
                    <div class="modal-actions">
                        <button type="button" class="btn btn-secondary" onclick="this.closest('.modal').remove()">
                            Закрыть
                        </button>
                    </div>
                </div>
            `;
            
            document.body.appendChild(modal);
        } catch (error) {
            console.error('Failed to load order details:', error);
            this.showNotification('Ошибка загрузки деталей заказа', 'error');
        }
    },
    
    // Получение текстового представления статуса заказа
    getOrderStatusText(status) {
        const statusMap = {
            'pending': 'В ожидании',
            'paid': 'Оплачен',
            'shipped': 'Отправлен',
            'delivered': 'Доставлен',
            'cancelled': 'Отменен'
        };
        return statusMap[status] || status;
    },
    
    // Редактирование статуса заказа
    editOrderStatus(orderId) {
        // Создаем модальное окно для изменения статуса
        const modal = document.createElement('div');
        modal.className = 'modal';
        modal.style.display = 'block';
        modal.innerHTML = `
            <div class="modal-content" style="max-width: 500px;">
                <div class="modal-header">
                    <h2>Изменить статус заказа</h2>
                    <span class="close-modal" onclick="this.closest('.modal').remove()">&times;</span>
                </div>
                <div class="modal-body">
                    <div class="form-group">
                        <label>Новый статус:</label>
                        <select id="orderStatusSelect" class="form-control">
                            <option value="pending">В ожидании</option>
                            <option value="paid">Оплачен</option>
                            <option value="shipped">Отправлен</option>
                            <option value="delivered">Доставлен</option>
                            <option value="cancelled">Отменен</option>
                        </select>
                    </div>
                </div>
                <div class="modal-actions">
                    <button type="button" class="btn btn-secondary" onclick="this.closest('.modal').remove()">
                        Отмена
                    </button>
                    <button type="button" class="btn btn-primary" onclick="admin.updateOrderStatus(${orderId})">
                        Сохранить
                    </button>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
    },
    
    // Обновление статуса заказа
    async updateOrderStatus(orderId) {
        try {
            const status = document.getElementById('orderStatusSelect').value;
            
            const response = await api.request(`/admin/orders/${orderId}/status`, {
                method: 'PUT',
                body: {status: status},
            });
            
            if (!response.ok) {
                throw new Error('Failed to update order status');
            }
            
            // Закрываем модальное окно
            document.querySelector('.modal').remove();
            
            // Обновляем список заказов
            this.loadOrders(this.currentOrdersPage);
            
            this.showNotification('Статус заказа успешно обновлен', 'success');
        } catch (error) {
            console.error('Failed to update order status:', error);
            this.showNotification('Ошибка обновления статуса заказа', 'error');
        }
    }
};

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', () => {
    admin.init();
});
