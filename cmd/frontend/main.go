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
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/redis"
	"github.com/gofiber/template/html/v2"
)

var sess *session.Store

func initRedis() *session.Store {
	// Создаем Redis storage для Fiber
	rdb := redis.New(redis.Config{
		Host:     "redis", //localhost
		Port:     6379,
		Password: "",
		Database: 0,
	})

	// Session store с Redis storage
	return session.New(session.Config{
		Storage:    rdb,
		KeyLookup:  "cookie:session_id",
		Expiration: 24 * time.Hour,
	})
}

func authMiddleware(c *fiber.Ctx) error {
	ses, err := sess.Get(c)
	if err != nil {
		return c.Next()
	}

	token := ses.Get("token")
	user := ses.Get("user") // ← interface{} (например, map[string]interface{})

	if token != nil {
		c.Locals("token", token)
		c.Locals("user", user) // ✅ Сохраняем в Locals
	}

	return c.Next()
}

// Пример использования в handler
// func someHandler(c *fiber.Ctx) error {
//     // Получаем сессию используя контекст Fiber
//     sess, err := sessions.Get(c)
//     if err != nil {
//         return c.Status(fiber.StatusInternalServerError).SendString("Session error")
//     }
//     defer sess.Save()

//     // Работа с сессией
//     sess.Set("user_id", 123)

//     return c.SendString("Hello World")
// }

func getEngineTemplate(webDir string, viewsLayout string) *fiber.App {
	// Инициализируем движок шаблонов
	engine := html.New(filepath.Join(webDir, "views"), ".html")
	engine.Reload(true) // Включить в dev
	return fiber.New(fiber.Config{
		DisableStartupMessage: false,
		Views:                 engine,
		ViewsLayout:           viewsLayout, // относительно views/
	})
}

func main() {
	// Находим папку web
	mTitle := "ODOBO - store"
	port0 := "3000"
	port := getEnv("PORT", "3001")
	urlDomaine := "http://localhost"
	urlStart := urlDomaine + ":" + port

	webDir := getWebDir()
	log.Println("📁 Using web directory:", webDir)

	app := getEngineTemplate(webDir, "layouts/main")
	// Middleware
	sess = initRedis()

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     urlDomaine + ":" + port0 + "," + urlStart,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// 1. Прокси API → ДО статики
	apiRoutes := []string{
		"/api/v1/auth/*",
		"/api/v1/products/*",
		"/api/v1/cart/*",
		"/api/v1/orders/*",
		"/api/v1/user/*",
		"/api/v1/admin/*",
	}
	setupAPIProxy(app, urlStart+"/", apiRoutes)

	// 2. Статические файлы
	setupStaticFiles(app, webDir)
	// 3. Страницы
	setupPageRoutes(app, mTitle)
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
			"Title": mTitle,
			"Page":  "app",
		})
	})

	// Запуск
	log.Println("🚀 Frontend server started on " + urlDomaine + ":" + port)
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
func setupAPIProxy(app *fiber.App, basetURL string, apiRoutes []string) {

	for _, route := range apiRoutes {
		app.All(route, func(c *fiber.Ctx) error {
			// Получаем путь после префикса
			path := c.Params("*")
			targetURL := basetURL + path

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

// getEnv — получить переменную окружения
func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func setupPageRoutes(app *fiber.App, mTitle string) {
	pages := map[string]struct {
		title string
		page  string
	}{
		"/":               {"Магазин модной женской одежды", "index"},
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
			app.Get(path, createDefaultHandler(config, mTitle))
		}
	}
}

func createDefaultHandler(config struct{ title, page string }, mTitle string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authMiddleware(c)
		user := c.Locals("user")
		return c.Render(config.page, fiber.Map{
			"Title": config.title + mTitle,
			"Page":  config.page,
			"User":  user,
		})
	}
}

func createProductsHandler(config struct{ title, page string }) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authMiddleware(c)
		category := c.Query("category")
		user := c.Locals("user")
		pageNum, _ := strconv.Atoi(c.Query("page", "1"))
		if pageNum < 1 {
			pageNum = 1
		}
		return c.Render(config.page, fiber.Map{
			"Title":       config.title + " - ODOBO store",
			"Page":        config.page,
			"Category":    category,
			"CurrentPage": pageNum,
			"User":        user,
		})
	}
}
