package handlers

import (
	"fmt"
	"otobo/internal/database/repositories"
	"otobo/internal/models"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ColorHandler struct {
	colorRepo *repositories.ColorRepository
}

func NewColorHandler(
	colorRepo *repositories.ColorRepository,
) *ColorHandler {
	return &ColorHandler{
		colorRepo: colorRepo,
	}
}

// GetColors возвращает все цвета
func (h *ColorHandler) GetColors(c *fiber.Ctx) error {
	colors, err := h.colorRepo.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch colors",
		})
	}

	return c.JSON(fiber.Map{
		"colors": colors,
	})
}

// GetActiveColors возвращает только активные цвета
func (h *ColorHandler) GetActiveColors(c *fiber.Ctx) error {
	colors, err := h.colorRepo.GetActive()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch active colors",
		})
	}

	return c.JSON(fiber.Map{
		"colors": colors,
	})
}

// GetColorByID возвращает цвет по ID
func (h *ColorHandler) GetColorByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	color, err := h.colorRepo.GetByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Color not found",
		})
	}

	return c.JSON(color)
}

// CreateColor создает новый цвет
func (h *ColorHandler) CreateColor(c *fiber.Ctx) error {
	// Логируем сырые данные
	body := c.Body()
	fmt.Printf("Raw body: %s\n", string(body))
	fmt.Printf("Content-Type: %s\n", c.Get("Content-Type"))
	var request struct {
		Name   string `json:"name"`
		Value  string `json:"value"`
		Active bool   `json:"active"`
	}
	if err := c.BodyParser(&request); err != nil {
		fmt.Printf("Parsed request: %+v\n", request)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Валидация
	if request.Name == "" || request.Value == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name and value are required",
		})
	}
	// Создаем цвет
	color := models.Color{
		Name:   request.Name,
		Value:  strings.ToUpper(request.Value),
		Active: request.Active,
	}
	if err := h.colorRepo.Create(&color); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create color",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(color)
}

// UpdateColor обновляет цвет
func (h *ColorHandler) UpdateColor(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	// Проверяем существование цвета
	// existingColor, err := h.colorRepo.GetByID(uint(id))
	// if err != nil {
	// 	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
	// 		"error": "Color not found",
	// 	})
	// }

	var color models.Color
	if err := c.BodyParser(&color); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Сохраняем ID из URL
	color.ID = uint(id)

	if err := h.colorRepo.Update(&color); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update color",
		})
	}

	return c.JSON(color)
}

// DeleteColor удаляет цвет
func (h *ColorHandler) DeleteColor(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	// Проверяем существование цвета
	_, err = h.colorRepo.GetByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Color not found",
		})
	}

	if err := h.colorRepo.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete color",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Color deleted successfully",
	})
}

// GetColorsByIDs возвращает цвета по списку ID
func (h *ColorHandler) GetColorsByIDs(c *fiber.Ctx) error {
	var request struct {
		IDs []uint `json:"ids"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	if len(request.IDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "IDs array is required",
		})
	}

	colors, err := h.colorRepo.GetByIDs(request.IDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch colors",
		})
	}

	return c.JSON(fiber.Map{
		"colors": colors,
	})
}
