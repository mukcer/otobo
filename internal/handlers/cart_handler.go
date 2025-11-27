package handlers

import (
	"otobo/internal/database/repositories"
	"time"

	"strconv"

	"github.com/gofiber/fiber/v2"
)

type CartHandler struct {
	cartRepo    *repositories.CartRepository
	productRepo *repositories.ProductRepository
}

func NewCartHandler(
	cartRepo *repositories.CartRepository,
	productRepo *repositories.ProductRepository,
) *CartHandler {
	return &CartHandler{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

// AddToCart добавляет товар в корзину
func (h *CartHandler) AddToCart(c *fiber.Ctx) error {
	var input struct {
		ProductID          uint `json:"product_id"`
		ProductVariationID uint `json:"variation_id"`
		Quantity           int  `json:"quantity"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	if input.Quantity <= 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Quantity must be greater than 0",
		})
	}

	userID, _ := c.Locals("userID").(uint)
	sessionID := c.Cookies("session_id")

	// Если пользователь не авторизован, создаем session_id
	if userID == 0 && sessionID == "" {
		sessionID = generateSessionID()
		c.Cookie(&fiber.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			HTTPOnly: true,
			Secure:   false, // true в production
		})
	}

	// Используем репозиторий для добавления в корзину
	cart, err := h.cartRepo.AddItem(userID, sessionID, input.ProductID, input.ProductVariationID, input.Quantity)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Product added to cart",
		"cart":    cart,
	})
}

// GetCart возвращает корзину пользователя
func (h *CartHandler) GetCart(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(uint)
	sessionID := c.Cookies("session_id")

	cart, err := h.cartRepo.GetCart(userID, sessionID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get cart",
		})
	}

	// Вычисляем общую стоимость
	total, err := h.cartRepo.CalculateCartTotal(userID, sessionID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to calculate cart total",
		})
	}

	return c.JSON(fiber.Map{
		"cart":       cart,
		"total":      total,
		"item_count": len(cart.Items),
	})
}

// UpdateCartItem обновляет количество товара в корзине
func (h *CartHandler) UpdateCartItem(c *fiber.Ctx) error {
	itemID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid item ID",
		})
	}

	var input struct {
		Quantity int `json:"quantity"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	userID, _ := c.Locals("userID").(uint)
	sessionID := c.Cookies("session_id")

	// Получаем корзину чтобы проверить принадлежность item
	cart, err := h.cartRepo.GetCart(userID, sessionID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Cart not found",
		})
	}

	// Проверяем, что item принадлежит корзине
	var itemExists bool
	for _, item := range cart.Items {
		if item.ID == uint(itemID) {
			itemExists = true
			break
		}
	}

	if !itemExists {
		return c.Status(404).JSON(fiber.Map{
			"error": "Item not found in cart",
		})
	}

	// Обновляем количество
	if err := h.cartRepo.UpdateItem(cart.ID, uint(itemID), input.Quantity); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Cart updated successfully",
	})
}

// RemoveFromCart удаляет товар из корзины
func (h *CartHandler) RemoveFromCart(c *fiber.Ctx) error {
	itemID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid item ID",
		})
	}

	if err := h.cartRepo.RemoveItem(uint(itemID)); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to remove item from cart",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Item removed from cart",
	})
}

// ClearCart очищает корзину
func (h *CartHandler) ClearCart(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(uint)
	sessionID := c.Cookies("session_id")

	cart, err := h.cartRepo.GetCart(userID, sessionID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Cart not found",
		})
	}

	if err := h.cartRepo.ClearCart(cart.ID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to clear cart",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Cart cleared successfully",
	})
}

// GetCartCount возвращает количество товаров в корзине
func (h *CartHandler) GetCartCount(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(uint)
	sessionID := c.Cookies("session_id")

	count, err := h.cartRepo.GetCartItemsCount(userID, sessionID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get cart count",
		})
	}

	return c.JSON(fiber.Map{
		"count": count,
	})
}

// generateSessionID генерирует уникальный session_id
func generateSessionID() string {
	return "session_" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

// GetCartByID возвращает корзину по ID
func (h *CartHandler) GetCartByID(c *fiber.Ctx) error {
	cartID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid cart ID",
		})
	}

	cart, err := h.cartRepo.GetByID(uint(cartID))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Cart not found",
		})
	}

	return c.JSON(fiber.Map{
		"cart": cart,
	})
}
