// handlers/auth.go
package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/valkey"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"otobo/internal/database/repositories"
	"otobo/internal/models"
)

type AuthHandler struct {
	userRepo    *repositories.UserRepository
	valkeyStore *valkey.Storage
	jwtSecret   string
}

func NewAuthHandler(
	userRepo *repositories.UserRepository,
	valkeyStore *valkey.Storage,
	jwtSecret string,
) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		valkeyStore: valkeyStore,
		jwtSecret:   jwtSecret,
	}
}

type LoginRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ClientTimestamp int64  `json:"client_timestamp"`
	HasLocalData    bool   `json:"has_local_data"`
}

type SessionData struct {
	UserID     string                 `json:"user_id"`
	LoginTime  string                 `json:"login_time"`
	UserAgent  string                 `json:"user_agent"`
	ClientData map[string]interface{} `json:"client_data"`
}

// 🔐 Логин
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	rawBody := c.Body()
	contentType := c.Get("Content-Type")
	log.Printf("Received Content-Type: %s\n", contentType) //
	err := json.Unmarshal(rawBody, &req)
	if err != nil {
		log.Printf("Ошибка Unmarshal JSON: %v. Сырые данные: %s\n", err, string(rawBody))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Неверный формат данных",
		})
	}
	// if err := c.BodyParser(&req); err != nil {
	// 	log.Printf("Ошибка парсинга тела запроса. Сырые данные: %s\n", string(rawBody))
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 		"error": "Неверный формат данных",
	// 	})
	// }

	// Валидация
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email и пароль обязательны",
		})
	}

	// Поиск пользователя через репозиторий
	user, err := h.userRepo.FindByEmail(strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Неверный email или пароль",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка сервера",
		})
	}

	// Проверка пароля
	if !user.CheckPassword(req.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Неверный email или пароль",
		})
	}

	// Генерация JWT токена
	token, err := h.generateJWT(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка генерации токена",
		})
	}

	// Создание сессии в Valkey
	sessionData := SessionData{
		UserID:    strconv.FormatUint(uint64(user.ID), 10),
		LoginTime: time.Now().Format(time.RFC3339),
		UserAgent: c.Get("User-Agent"),
		ClientData: map[string]any{
			"last_sync":      req.ClientTimestamp,
			"has_local_data": req.HasLocalData,
			"timezone":       "Europe/Moscow",
		},
	}

	if err := h.saveSessionValkey(strconv.FormatUint(uint64(user.ID), 10), sessionData); err != nil {
		fmt.Printf("Failed to save session to Valkey: %v\n", err)
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user":  user.ToResponse(),
	})
}

// 📝 Регистрация
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Неверный формат данных при регистрации",
		})
	}

	// Валидация
	if strings.TrimSpace(req.FirstName) == "" ||
		strings.TrimSpace(req.LastName) == "" ||
		strings.TrimSpace(req.Email) == "" ||
		strings.TrimSpace(req.Password) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Все обязательные поля должны быть заполнены",
		})
	}

	// Проверка существования пользователя через репозиторий
	existingUser, err := h.userRepo.FindByEmail(strings.ToLower(strings.TrimSpace(req.Email)))
	if err == nil && existingUser != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Пользователь с таким email уже существует",
		})
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка сервера",
		})
	}

	// Создание пользователя
	user := &models.User{
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
		Email:     strings.ToLower(strings.TrimSpace(req.Email)),
		Phone:     strings.TrimSpace(req.Phone),
		Address:   strings.TrimSpace(req.Address),
		Password:  req.Password,
	}

	// Проверяем количество существующих пользователей
	userCount, err := h.userRepo.Count()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка сервера",
		})
	}

	// Если это первый пользователь, назначаем ему роль admin
	if userCount == 0 {
		user.Role = "admin"
	} else {
		user.Role = "customer"
	}

	// Сохранение через репозиторий
	if err := h.userRepo.Create(user); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Пользователь с таким email уже существует",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка создания пользователя",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Пользователь успешно зарегистрирован",
		"user":    user.ToResponse(),
	})
}

// 🔄 Синхронизация
func (h *AuthHandler) Sync(c *fiber.Ctx) error {
	userIDStr := c.Locals("userID").(string)

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Неверный ID пользователя",
		})
	}

	user, err := h.userRepo.FindByID(uint(userID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Пользователь не найден",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка сервера",
		})
	}

	// Обновление сессии в Valkey
	sessionKey := "session:" + userIDStr
	now := time.Now().Format(time.RFC3339)

	// Получаем текущие данные сессии
	var sessionData SessionData
	storedData, err := h.valkeyStore.Get(sessionKey)
	if err == nil && storedData != nil {
		if err := json.Unmarshal(storedData, &sessionData); err != nil {
			fmt.Printf("Failed to unmarshal session data: %v\n", err)
			// Создаем новую сессию если данные повреждены
			sessionData = SessionData{
				UserID:     userIDStr,
				LoginTime:  now,
				UserAgent:  c.Get("User-Agent"),
				ClientData: make(map[string]interface{}),
			}
		}
	} else {
		// Создаем новую сессию если не найдена
		sessionData = SessionData{
			UserID:     userIDStr,
			LoginTime:  now,
			UserAgent:  c.Get("User-Agent"),
			ClientData: make(map[string]interface{}),
		}
	}

	// Обновляем поля
	sessionData.ClientData["last_sync"] = now
	sessionData.ClientData["last_active"] = now

	// Сохраняем обновленные данные
	if err := h.saveSessionValkey(userIDStr, sessionData); err != nil {
		fmt.Printf("Failed to update session sync time: %v\n", err)
	}

	return c.JSON(fiber.Map{
		"user":      user.ToResponse(),
		"synced_at": now,
	})
}

// 👤 Получение профиля
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	userIDStr := c.Locals("userID").(string)

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Неверный ID пользователя",
		})
	}

	user, err := h.userRepo.FindByID(uint(userID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Пользователь не найден",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка сервера",
		})
	}

	return c.JSON(fiber.Map{
		"user": user.ToResponse(),
	})
}

// 🔐 Создание сессии
func (h *AuthHandler) CreateSession(c *fiber.Ctx) error {
	var req SessionData
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Неверный формат данных при создании сессии",
		})
	}

	userID, err := strconv.ParseUint(req.UserID, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Неверный ID пользователя",
		})
	}

	_, err = h.userRepo.FindByID(uint(userID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Пользователь не найден",
		})
	}

	if err := h.saveSessionValkey(req.UserID, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка создания сессии",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "session_created",
		"user_id": req.UserID,
	})
}

// 🚪 Выход
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	sessionKey := "session:" + userID
	if err := h.valkeyStore.Delete(sessionKey); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка выхода",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Выход выполнен успешно",
	})
}

// 🛡️ Middleware аутентификации
func (h *AuthHandler) AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Требуется авторизация",
		})
	}

	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method)
		}

		return []byte(h.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Неверный токен",
		})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Неверные данные токена",
		})
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Неверный ID пользователя в токене",
		})
	}

	// Проверка сессии в Valkey через Get
	sessionKey := "session:" + userID
	storedData, err := h.valkeyStore.Get(sessionKey)
	if err != nil || storedData == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Сессия истекла",
		})
	}

	c.Locals("userID", userID)
	c.Locals("userRole", claims["role"])
	return c.Next()
}

// 🔧 Вспомогательные методы

func (h *AuthHandler) generateJWT(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": strconv.FormatUint(uint64(user.ID), 10),
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}
func (h *AuthHandler) AdminMiddleware(c *fiber.Ctx) error {
	// This assumes AuthAPIMiddleware ran before this
	userRole, ok := c.Locals("userRole").(string)
	if !ok || userRole != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{ // Use 403 Forbidden
			"error": "Access denied. Admin rights required",
		})
	}

	return c.Next()
}
func (h *AuthHandler) saveSessionValkey(userID string, sessionData SessionData) error {
	sessionKey := "session:" + userID
	sessionJSON, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}
	duration := 7 * 24 * time.Hour
	if err := h.valkeyStore.Set(sessionKey, sessionJSON, duration); err != nil {
		return fmt.Errorf("failed to save session to valkey: %w", err)
	}
	return nil
}

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
