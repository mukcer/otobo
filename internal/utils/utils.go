// getEnv — получить переменную окружения
package utils

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
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
	engine.Reload(true) // Включить в dev
	return fiber.New(fiber.Config{
		DisableStartupMessage: false,
		Views:                 engine,
		ViewsLayout:           viewsLayout, // относительно views/
	})
}
