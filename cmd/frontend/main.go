package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/template/html/v2"
)

func main() {
	// Находим папку web
	webDir := getWebDir()
	log.Println("📁 Using web directory:", webDir)

	// Инициализируем движок шаблонов
	engine := html.New(filepath.Join(webDir, "views"), ".html")
	engine.Reload(true) // Включить в dev

	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
		Views:                 engine,
		ViewsLayout:           "layouts/main", // относительно views/
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,http://localhost:3001",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// === ВАЖНО: Порядок роутов ===

	// 1. Прокси API → ДО статики
	setupAPIProxy(app)

	// 2. Статические файлы
	// У вас: web/css, web/js, web/images
	app.Static("/css", filepath.Join(webDir, "css"), fiber.Static{
    	CacheDuration: -1, // Отключает кэширование 
	})
	app.Static("/js", filepath.Join(webDir, "js"), fiber.Static{
    	CacheDuration: -1, // Отключает кэширование 
	})
	app.Static("/images", filepath.Join(webDir, "images"))
	app.Static("/static", webDir) // резервный путь, если где-то /static/...

	// 3. Страницы
	setupPageRoutes(app)

	// 4. Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "healthy",
			"service":   "frontend",
			"timestamp": time.Now(),
		})
	})

	// 5. SPA Fallback — последний
	app.Use(func(c *fiber.Ctx) error {
		path := c.Path()

		// Пропускаем API, статику, .ico и т.п.
		if strings.HasPrefix(path, "/api/") ||
			strings.Contains(path, ".") ||
			strings.HasPrefix(path, "/css/") ||
			strings.HasPrefix(path, "/js/") ||
			strings.HasPrefix(path, "/images/") {
			return c.SendStatus(404)
		}

		return c.Render("index", fiber.Map{
			"Title": "Fashion Store",
			"Page":  "app",
		})
	})

	// Запуск
	port := getEnv("PORT", "3001")
	log.Println("🚀 Frontend server started on http://localhost:" + port)
	log.Fatal(app.Listen(":" + port))
}

// getWebDir — ищем папку web
func getWebDir() string {
	currentDir, _ := os.Getwd()
	log.Println("🔍 Current dir:", currentDir)

	// Относительные пути от cmd/frontend
	dirsToCheck := []string{
		filepath.Join(currentDir, "..", "..", "web"), // ../../web
		filepath.Join(currentDir, "..", "web"),       // ../web
		filepath.Join(currentDir, "web"),             // ./web
		"../../web",
		"../web",
		"./web",
	}

	for _, dir := range dirsToCheck {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			abs, _ := filepath.Abs(dir)
			return abs
		}
	}

	// Если не найдено — паника
	log.Fatal("❌ Папка 'web' не найдена. Ожидается: ../../web")
	return ""
}

// setupAPIProxy — проксируем API на бэкенд (:3000)
func setupAPIProxy(app *fiber.App) {
	apiRoutes := []string{
		"/api/v1/auth/*",
		"/api/v1/products/*",
		"/api/v1/cart/*",
		"/api/v1/orders/*",
		"/api/v1/user/*",
		"/api/v1/admin/*",
	}

	for _, route := range apiRoutes {
		app.All(route, func(c *fiber.Ctx) error {
			// Получаем путь после префикса
			path := c.Params("*")
			targetURL := "http://localhost:3000/api/v1/" + path

			// Выполняем прокси
			if err := proxy.Do(c, targetURL); err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": "API server is unreachable",
				})
			}

			// Убираем заголовок Server
			c.Response().Header.Del(fiber.HeaderServer)
			return nil
		})
	}
}

// setupPageRoutes — страницы
func setupPageRoutes(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title": "Fashion Store - Магазин модной женской одежды",
			"Page":  "home",
		})
	})

	app.Get("/products", func(c *fiber.Ctx) error {
		category := c.Query("category")
		page, _ := strconv.Atoi(c.Query("page", "1"))
		if page < 1 {
			page = 1
		}

		return c.Render("products", fiber.Map{
			"Title":       "Каталог - Fashion Store",
			"Page":        "products",
			"Category":    category,
			"CurrentPage": page,
		})
	})

	app.Get("/login", func(c *fiber.Ctx) error {
		return c.Render("login", fiber.Map{
			"Title": "Вход - Fashion Store",
			"Page":  "login",
		})
	})

	app.Get("/register", func(c *fiber.Ctx) error {
		return c.Render("register", fiber.Map{
			"Title": "Регистрация - Fashion Store",
			"Page":  "register",
		})
	})

	app.Get("/profile", func(c *fiber.Ctx) error {
		return c.Render("profile", fiber.Map{
			"Title": "Профиль - Fashion Store",
			"Page":  "profile",
		})
	})

	app.Get("/admin/products", func(c *fiber.Ctx) error {
		return c.Render("admin_products", fiber.Map{
			"Title": "Управление товарами",
			"Page":  "admin_products",
		})
	})

	app.Get("/cart", func(c *fiber.Ctx) error {
		return c.Render("cart", fiber.Map{
			"Title": "Корзина - Fashion Store",
			"Page":  "cart",
		})
	})
}

// getEnv — получить переменную окружения
func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
