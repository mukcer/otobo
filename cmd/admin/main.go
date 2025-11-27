package main

import (
	"log"
	"os"
	"otobo/internal/utils"
	"otobo/internal/weinkey"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/fiber/v2/middleware/session"
)

var sess *session.Store

func main() {
	mTitle := "ODOBO - admin"
	apiRoutes := []string{
		"/api/v1/auth/*",
		"/api/v1/products/*",
		"/api/v1/cart/*",
		"/api/v1/orders/*",
		"/api/v1/user/*",
		"/api/v1/admin/*",
		"/api/v1/colors/*",
	}
	port := utils.GetEnv("ADMIN_PORT", "3002")
	store := weinkey.ValkeyInit()
	sess = weinkey.SessionInit(store)
	apiBaseURL := utils.GetEnv("API_URL", "http://localhost:3000")
	mainInit(apiBaseURL, apiRoutes, mTitle, port)
}
func setupPageRoutes(app *fiber.App, mTitle string) {
	pages := map[string]struct {
		title string
		page  string
	}{
		"/":           {"Панель управления", "admin"},
		"/products":   {"Управление товарами", "admin_products"},
		"/categories": {"Управление категориями", "admin_categories"},
		"/login":      {"Вход", "login"},
		"/register":   {"Регистрация", "register"},
		"/profile":    {"Профиль", "profile"},
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
			"Title":       config.title + " - ODOBO Admin",
			"Page":        config.page,
			"Category":    category,
			"CurrentPage": pageNum,
			"User":        user,
		})
	}
}

func authMiddleware(c *fiber.Ctx) error {
	ses, err := sess.Get(c)
	if err != nil {
		return c.Next()
	}

	token := ses.Get("auth_token")
	user := ses.Get("user_data") // ← interface{} (например, map[string]interface{})

	if token != nil {
		c.Locals("token", token)
		c.Locals("user", user) // ✅ Сохраняем в Locals
	}

	return c.Next()
}

func mainInit(apiBaseURL string, apiRoutes []string, mTitle string, port string) *fiber.App {
	port0 := "3000"
	urlDomaine := "http://localhost"
	urlStart := urlDomaine + ":" + port
	webDir := utils.GetWebDir()
	app := utils.GetEngineTemplate(webDir, "layouts/main")

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     urlDomaine + ":" + port0 + "," + urlStart,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	setupAPIProxy(app, apiBaseURL+"/", apiRoutes)
	setupStaticFiles(app, webDir)
	log.Println("📁 Using web directory:", webDir)
	setupPageRoutes(app, mTitle)
	// 4. Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "healthy",
			"service":   "admin",
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
	log.Println("🚀 Admin server started on " + urlDomaine + ":" + port)
	log.Printf("📁 Web directory: %s", webDir)
	log.Fatal(app.Listen(":" + port))
	return app
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
