package handlers

import (
	"otobo/internal/database/repositories"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type OrderHandler struct {
	orderRepo *repositories.OrderRepository
	cartRepo  *repositories.CartRepository
}

func NewOrderHandler(
	orderRepo *repositories.OrderRepository,
	cartRepo *repositories.CartRepository,
) *OrderHandler {
	return &OrderHandler{
		orderRepo: orderRepo,
		cartRepo:  cartRepo,
	}
}

func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	var input struct {
		ShippingAddress string `json:"shipping_address"`
		ShippingMethod  string `json:"shipping_method"`
		PaymentMethod   string `json:"payment_method"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Создаем заказ через репозиторий
	order, err := h.orderRepo.CreateFromCart(userID, input.ShippingAddress, input.ShippingMethod, input.PaymentMethod)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Order created successfully",
		"order":   order,
	})
}

func (h *OrderHandler) GetUserOrders(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	orders, err := h.orderRepo.FindByUserID(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch orders",
		})
	}

	return c.JSON(fiber.Map{
		"orders": orders,
	})
}

// GetOrder возвращает конкретный заказ по ID
func (h *OrderHandler) GetOrder(c *fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid order ID",
		})
	}

	order, err := h.orderRepo.GetOrderByID(uint(orderID))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Order not found",
		})
	}

	// Проверяем, что заказ принадлежит пользователю
	userID := c.Locals("userID").(uint)
	if order.UserID != userID {
		return c.Status(403).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	return c.JSON(fiber.Map{
		"order": order,
	})
}

// UpdateOrderStatus обновляет статус заказа (для администратора)
func (h *OrderHandler) UpdateOrderStatus(c *fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid order ID",
		})
	}

	var input struct {
		Status string `json:"status"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Проверяем, что статус допустимый
	validStatuses := map[string]bool{
		"pending":   true,
		"paid":      true,
		"shipped":   true,
		"delivered": true,
		"cancelled": true,
	}

	if !validStatuses[input.Status] {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid status",
		})
	}

	if err := h.orderRepo.UpdateOrderStatus(uint(orderID), input.Status); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update order status",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Order status updated successfully",
	})
}
