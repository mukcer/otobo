// getEnv — получить переменную окружения
package utils

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/template/html/v2"
)

func GetEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

// getWebDir — ищем папку web
func GetWebDir() string {
	currentDir, _ := os.Getwd()
	log.Println("🔍 Current dir:", currentDir)

	// Относительные пути от cmd/admin
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
func GetEngineTemplate(webDir string, viewsLayout string) *fiber.App {
	// Инициализируем движок шаблонов
	engine := html.New(filepath.Join(webDir, "views"), ".html")
	engine.Reload(true)
	engine.Debug(true) // Включить в dev
	return fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: viewsLayout, // относительно views/
	})
}

// setupAPIProxy — проксируем API на бэкенд (:3000)
func SetupAPIProxy(app *fiber.App, basetURL string, apiRoutes []string) {

	for _, route := range apiRoutes {
		app.All(route, func(c *fiber.Ctx) error {
			// Получаем путь после префикса
			path := c.Params("*")
			targetURL := basetURL + path

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
func SetupStaticFiles(app *fiber.App, webDir string) {
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
