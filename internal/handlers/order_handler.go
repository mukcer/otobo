package handlers

import (
	"otobo/internal/database/repositories"

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
