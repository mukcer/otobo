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
	setupStaticFiles(app, webDir)

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
	setupSPAFallback(app)

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
	log.Printf("📁 Web directory: %s", webDir)
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
			targetURL := "http://localhost:3000/" + path

			// Выполняем прокси
			if err := proxy.Do(c, targetURL); err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": "API server is unreachable",
				})
			}

			if err := proxy.DoTimeout(c, targetURL, 10*time.Second); err != nil {
				log.Printf("❌ Proxy error: %v", err)
				return c.Status(502).JSON(fiber.Map{
					"error": "API server is unreachable",
				})
			}

			// Убираем заголовок Server
			c.Response().Header.Del(fiber.HeaderServer)
			return nil
		})
	}
}
func setupStaticFiles(app *fiber.App, webDir string) {
	// Для разработки отключаем кэширование
	cacheDuration := -1 * time.Second
	if os.Getenv("APP_ENV") == "production" {
		cacheDuration = 24 * time.Hour
	}

	app.Static("/css", filepath.Join(webDir, "css"), fiber.Static{
		CacheDuration: cacheDuration,
		MaxAge:        int(cacheDuration.Seconds()),
	})

	app.Static("/js", filepath.Join(webDir, "js"), fiber.Static{
		CacheDuration: cacheDuration,
		MaxAge:        int(cacheDuration.Seconds()),
	})

	app.Static("/images", filepath.Join(webDir, "images"), fiber.Static{
		CacheDuration: cacheDuration,
	})

	app.Static("/static", webDir, fiber.Static{
		CacheDuration: cacheDuration,
	})
}

// setupPageRoutes — страницы
func setupSPAFallback(app *fiber.App) {
	app.Use(func(c *fiber.Ctx) error {
		path := c.Path()

		// Игнорируем API, статику и файлы с расширениями
		if strings.HasPrefix(path, "/api/") ||
			strings.HasPrefix(path, "/css/") ||
			strings.HasPrefix(path, "/js/") ||
			strings.HasPrefix(path, "/images/") ||
			strings.HasPrefix(path, "/static/") ||
			strings.Contains(path, ".") {
			return c.SendStatus(404)
		}

		// Для всех остальных маршрутов отдаем SPA
		return c.Render("index", fiber.Map{
			"Title": "ODOBO Store",
			"Page":  "app",
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

func setupPageRoutes(app *fiber.App) {
	pages := map[string]struct {
		title string
		page  string
	}{
		"/":               {"Магазин модной женской одежды", "home"},
		"/products":       {"Каталог", "products"},
		"/login":          {"Вход", "login"},
		"/register":       {"Регистрация", "register"},
		"/profile":        {"Профиль", "profile"},
		"/admin/products": {"Управление товарами", "admin_products"},
		"/cart":           {"Корзина", "cart"},
	}

	for path, config := range pages {
		if path == "/products" {
			app.Get(path, createProductsHandler(config))
		} else {
			app.Get(path, createDefaultHandler(config))
		}
	}
}

func createDefaultHandler(config struct{ title, page string }) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title": config.title + " - ODOBO store",
			"Page":  config.page,
		})
	}
}

func createProductsHandler(config struct{ title, page string }) fiber.Handler {
	return func(c *fiber.Ctx) error {
		category := c.Query("category")
		pageNum, _ := strconv.Atoi(c.Query("page", "1"))
		if pageNum < 1 {
			pageNum = 1
		}
		return c.Render("index", fiber.Map{
			"Title":       config.title + " - ODOBO store",
			"Page":        config.page,
			"Category":    category,
			"CurrentPage": pageNum,
		})
	}
}
