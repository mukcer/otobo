// handlers/auth.go
package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"otobo/internal/database/repositories"
	"otobo/internal/models"
)

type AuthHandler struct {
	userRepo    *repositories.UserRepository
	redisClient *redis.Client
	jwtSecret   string
}

func NewAuthHandler(
	userRepo *repositories.UserRepository,
	redisClient *redis.Client,
	jwtSecret string,
) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		redisClient: redisClient,
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
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Неверный формат данных",
		})
	}

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

	// Создание сессии в Redis
	sessionData := SessionData{
		UserID:    strconv.FormatUint(uint64(user.ID), 10),
		LoginTime: time.Now().Format(time.RFC3339),
		UserAgent: c.Get("User-Agent"),
		ClientData: map[string]interface{}{
			"last_sync":      req.ClientTimestamp,
			"has_local_data": req.HasLocalData,
		},
	}

	if err := h.saveSessionToRedis(c, strconv.FormatUint(uint64(user.ID), 10), sessionData); err != nil {
		// Логируем ошибку, но не прерываем процесс
		fmt.Printf("Failed to save session to Redis: %v\n", err)
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
			"error": "Неверный формат данных",
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
		Password:  req.Password, // Пароль автоматически хешируется в BeforeSave
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

	// Конвертация userID из string в uint
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Неверный ID пользователя",
		})
	}

	// Получение пользователя через репозиторий
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

	// Обновление сессии в Redis
	sessionKey := "session:" + userIDStr
	if err := h.redisClient.HSet(c.Context(), sessionKey,
		"last_sync", time.Now().Format(time.RFC3339),
		"last_active", time.Now().Format(time.RFC3339),
	).Err(); err != nil {
		fmt.Printf("Failed to update session sync time: %v\n", err)
	}

	return c.JSON(fiber.Map{
		"user":      user.ToResponse(),
		"synced_at": time.Now().Format(time.RFC3339),
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
			"error": "Неверный формат данных",
		})
	}

	// Проверяем существование пользователя через репозиторий
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

	if err := h.saveSessionToRedis(c, req.UserID, req); err != nil {
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

	// Удаление сессии из Redis
	sessionKey := "session:" + userID
	if err := h.redisClient.Del(c.Context(), sessionKey).Err(); err != nil {
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

	// Извлечение токена из заголовка
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

	// Валидация JWT токена
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
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

	// Извлечение userID из токена
	userID, ok := claims["user_id"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Неверный ID пользователя в токене",
		})
	}

	// Проверка сессии в Redis
	sessionKey := "session:" + userID
	exists, err := h.redisClient.Exists(c.Context(), sessionKey).Result()
	if err != nil || exists == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Сессия истекла",
		})
	}

	// Сохранение userID в контексте
	c.Locals("userID", userID)

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

func (h *AuthHandler) saveSessionToRedis(c *fiber.Ctx, userID string, sessionData SessionData) error {
	sessionKey := "session:" + userID

	data := map[string]interface{}{
		"user_id":     sessionData.UserID,
		"login_time":  sessionData.LoginTime,
		"user_agent":  sessionData.UserAgent,
		"client_data": sessionData.ClientData,
		"created_at":  time.Now().Format(time.RFC3339),
		"last_active": time.Now().Format(time.RFC3339),
	}

	if err := h.redisClient.HSet(c.Context(), sessionKey, data).Err(); err != nil {
		return err
	}

	// Установка TTL 24 часа
	return h.redisClient.Expire(c.Context(), sessionKey, 24*time.Hour).Err()
}
