package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// FrontendHandler имеет зависимость только от session store
type FrontendHandler struct {
	sessionStore *session.Store
}

// NewFrontendHandler — конструктор для фронтенд-обработчика
func NewFrontendHandler(store *session.Store) *FrontendHandler {
	return &FrontendHandler{
		sessionStore: store,
	}
}
func (h *FrontendHandler) SessionAuthMiddleware(c *fiber.Ctx) error {
	ses, err := h.sessionStore.Get(c)
	if err != nil {
		log.Printf("Session store error: %v", err)
		return c.Next() // Разрешаем идти дальше, просто как анонимный пользователь
	}

	token := ses.Get("auth_token")
	user := ses.Get("user_data")

	if token != nil && user != nil {
		c.Locals("user", user) // ✅ Сохраняем в Locals для использования в шаблонах
	}

	return c.Next()
}
