package handlers

import (
	"strconv"

	"otobo/internal/database/repositories"

	"github.com/gofiber/fiber/v2"
)

type ProductHandler struct {
	productRepo  *repositories.ProductRepository
	categoryRepo *repositories.CategoryRepository
}

func NewProductHandler(
	productRepo *repositories.ProductRepository,
	categoryRepo *repositories.CategoryRepository,
) *ProductHandler {
	return &ProductHandler{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

func (h *ProductHandler) GetProducts(c *fiber.Ctx) error {
	// Парсим параметры запроса
	filter := repositories.ProductFilter{
		CategorySlug: c.Query("category"),
		Size:         c.Query("size"),
		Color:        c.Query("color"),
		InStock:      c.Query("in_stock") == "true",
		IsActive:     true, // По умолчанию только активные товары
	}

	if minPrice := c.Query("min_price"); minPrice != "" {
		if min, err := strconv.ParseFloat(minPrice, 64); err == nil {
			filter.MinPrice = min
		}
	}

	if maxPrice := c.Query("max_price"); maxPrice != "" {
		if max, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			filter.MaxPrice = max
		}
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "12"))
	sortBy := c.Query("sort", "created_at")
	order := c.Query("order", "desc")

	// Используем репозиторий
	products, total, err := h.productRepo.FindAll(filter, page, limit, sortBy, order)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to fetch products",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"products": products,
		"total":    total,
		"page":     page,
		"limit":    limit,
		"pages":    (int(total) + limit - 1) / limit,
	})
}

func (h *ProductHandler) GetProduct(c *fiber.Ctx) error {
	slug := c.Params("slug")

	product, err := h.productRepo.FindBySlug(slug)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Product not found",
		})
	}

	return c.JSON(product)
}

func (h *ProductHandler) GetCategories(c *fiber.Ctx) error {
	categories, err := h.categoryRepo.FindAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch categories",
		})
	}

	return c.JSON(categories)
}

func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	// Логика создания товара через репозиторий
	// ...
	return c.JSON(fiber.Map{"message": "Product created"})
}
