package handlers

import (
	"log"
	"otobo/internal/database"
	"otobo/internal/database/repositories"
	"otobo/internal/weinkey"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/valkey"
)

type AdminHandler struct {
	db    *database.Database
	store *valkey.Storage
}

func NewAdminHandler(db *database.Database, store *valkey.Storage) *AdminHandler {
	return &AdminHandler{
		db:    db,
		store: store,
	}
}

// GetDashboardStats возвращает статистику для дашборда админки
func (h *AdminHandler) GetDashboardStats(c *fiber.Ctx) error {
	// Здесь будет логика получения статистики
	stats := map[string]interface{}{
		"total_products": 142,
		"total_users":    1248,
		"total_orders":   89,
		"total_revenue":  124560,
	}

	return c.JSON(stats)
}

// GetDatabaseTables возвращает список таблиц в базе данных
func (h *AdminHandler) GetDatabaseTables(c *fiber.Ctx) error {
	tables, err := h.db.GetTables()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get tables",
		})
	}

	// Получаем дополнительную информацию о каждой таблице
	tableInfo := make([]map[string]interface{}, 0)
	for _, table := range tables {
		info, err := h.db.GetTableInfo(table)
		if err != nil {
			log.Printf("Failed to get info for table %s: %v", table, err)
			continue
		}
		tableInfo = append(tableInfo, info)
	}

	return c.JSON(tableInfo)
}

// GetTableData возвращает данные из таблицы
func (h *AdminHandler) GetTableData(c *fiber.Ctx) error {
	tableName := c.Params("tableName")
	if tableName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Table name is required",
		})
	}

	// Параметры пагинации
	limit := c.QueryInt("limit", 10)
	offset := c.QueryInt("offset", 0)

	// Получаем данные из таблицы
	data, err := h.db.GetTableData(tableName, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get table data",
		})
	}

	// Получаем информацию о колонках
	columns, err := h.db.GetTableColumns(tableName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get table columns",
		})
	}

	return c.JSON(fiber.Map{
		"data":    data,
		"columns": columns,
	})
}

// UpdateTableData обновляет данные в таблице
func (h *AdminHandler) UpdateTableData(c *fiber.Ctx) error {
	tableName := c.Params("tableName")
	if tableName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Table name is required",
		})
	}

	// Получаем ID записи
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Record ID is required",
		})
	}

	// Парсим данные из тела запроса
	var data map[string]interface{}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Обновляем данные в таблице
	err := h.db.UpdateTableData(tableName, id, data)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update table data",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Record updated successfully",
	})
}

// DeleteDatabaseTable удаляет таблицу из базы данных
func (h *AdminHandler) DeleteDatabaseTable(c *fiber.Ctx) error {
	tableName := c.Params("tableName")
	if tableName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Table name is required",
		})
	}

	// Защита от удаления системных таблиц
	protectedTables := []string{"users", "products", "orders", "categories", "colors", "sizes", "product_variations", "carts", "cart_items", "order_items", "reviews"}
	for _, protected := range protectedTables {
		if tableName == protected {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Cannot delete protected table",
			})
		}
	}

	err := h.db.DeleteTable(tableName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete table",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Table deleted successfully",
	})
}

// BackupDatabase создает резервную копию базы данных
func (h *AdminHandler) BackupDatabase(c *fiber.Ctx) error {
	err := h.db.BackupDatabase()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create backup",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Database backup created successfully",
	})
}

// OptimizeDatabase оптимизирует базу данных
func (h *AdminHandler) OptimizeDatabase(c *fiber.Ctx) error {
	err := h.db.OptimizeDatabase()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to optimize database",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Database optimized successfully",
	})
}

// ClearQueryCache очищает кэш запросов
func (h *AdminHandler) ClearQueryCache(c *fiber.Ctx) error {
	err := h.db.ClearQueryCache()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to clear query cache",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Query cache cleared successfully",
	})
}

// GetCacheStats возвращает статистику использования кэша
func (h *AdminHandler) GetCacheStats(c *fiber.Ctx) error {
	// Создаем клиент для управления Valkey
	valkeyClient := weinkey.NewAdminValkeyClient(h.store)

	stats, err := valkeyClient.GetStats()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get cache stats",
		})
	}

	return c.JSON(stats)
}

// GetCacheKeys возвращает список ключей в кэше
func (h *AdminHandler) GetCacheKeys(c *fiber.Ctx) error {
	// Создаем клиент для управления Valkey
	valkeyClient := weinkey.NewAdminValkeyClient(h.store)

	keys, err := valkeyClient.GetKeys()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get cache keys",
		})
	}

	return c.JSON(keys)
}

// DeleteCacheKey удаляет ключ из кэша
func (h *AdminHandler) DeleteCacheKey(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Key is required",
		})
	}

	// Создаем клиент для управления Valkey
	valkeyClient := weinkey.NewAdminValkeyClient(h.store)

	err := valkeyClient.DeleteKey(key)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete cache key",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Cache key deleted successfully",
	})
}

// ClearAllCache очищает весь кэш
func (h *AdminHandler) ClearAllCache(c *fiber.Ctx) error {
	// Создаем клиент для управления Valkey
	valkeyClient := weinkey.NewAdminValkeyClient(h.store)

	err := valkeyClient.ClearAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to clear cache",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Cache cleared successfully",
	})
}

// ClearSessions очищает все сессии
func (h *AdminHandler) ClearSessions(c *fiber.Ctx) error {
	// Создаем клиент для управления Valkey
	valkeyClient := weinkey.NewAdminValkeyClient(h.store)

	err := valkeyClient.ClearSessions()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to clear sessions",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Sessions cleared successfully",
	})
}

// GetOrders возвращает список заказов с фильтрацией и пагинацией
func (h *AdminHandler) GetOrders(c *fiber.Ctx) error {
	// Параметры фильтрации
	status := c.Query("status")
	userID := c.QueryInt("user_id", 0)
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	// Создаем фильтр
	filter := repositories.OrderFilter{
		Status: status,
		UserID: uint(userID),
		Page:   page,
		Limit:  limit,
	}

	// Получаем заказы через репозиторий
	orderRepo := repositories.NewOrderRepository(h.db.DB)
	orders, total, err := orderRepo.FindAll(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get orders",
		})
	}

	return c.JSON(fiber.Map{
		"orders": orders,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

// GetOrderDetails возвращает детали конкретного заказа
func (h *AdminHandler) GetOrderDetails(c *fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid order ID",
		})
	}

	// Получаем заказ через репозиторий
	orderRepo := repositories.NewOrderRepository(h.db.DB)
	order, err := orderRepo.GetOrderByID(uint(orderID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Order not found",
		})
	}

	return c.JSON(order)
}

// UpdateOrderStatus обновляет статус заказа
func (h *AdminHandler) UpdateOrderStatus(c *fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid order ID",
		})
	}

	var input struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Проверяем допустимые статусы
	validStatuses := map[string]bool{
		"pending":   true,
		"paid":      true,
		"shipped":   true,
		"delivered": true,
		"cancelled": true,
	}
	if !validStatuses[input.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid status",
		})
	}

	// Обновляем статус через репозиторий
	orderRepo := repositories.NewOrderRepository(h.db.DB)
	if err := orderRepo.UpdateOrderStatus(uint(orderID), input.Status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update order status",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Order status updated successfully",
	})
}
