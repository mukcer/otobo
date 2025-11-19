package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(401).JSON(fiber.Map{
			"error": "Authorization header required",
		})
	}

	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("your-jwt-secret"), nil // Должно быть из конфига
	})

	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}

	claims := token.Claims.(jwt.MapClaims)
	c.Locals("userID", uint(claims["user_id"].(float64)))
	c.Locals("userRole", claims["role"])

	return c.Next()
}

func AdminMiddleware(c *fiber.Ctx) error {
	userRole := c.Locals("userRole").(string)
	if userRole != "admin" {
		return c.Status(403).JSON(fiber.Map{
			"error": "Access denied. Admin rights required",
		})
	}

	return c.Next()
}
