package main

import (
	"log"
	"otobo/internal/handlers"
	"otobo/internal/utils"
	"otobo/internal/weinkey"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
)

var sess *session.Store

func main() {
	mTitle := "ODOBO - store"
	apiRoutes := []string{
		"/api/v1/auth/*",
		"/api/v1/products/*",
		"/api/v1/cart/*",
		"/api/v1/orders/*",
		"/api/v1/user/*",
		"/api/v1/admin/*",
		"/api/v1/colors/*",
	}
	port := utils.GetEnv("PORT", "3001")
	store := weinkey.ValkeyInit()
	sess = weinkey.SessionInit(store)
	apiBaseURL := utils.GetEnv("API_URL", "http://localhost:3000")
	mainInit(apiBaseURL, apiRoutes, mTitle, port)
}
func setupPageRoutes(app *fiber.App, mTitle string) {
	pages := map[string]handlers.PageHandler{
		"/":         {Title: "Магазин модной женской одежды", Page: "index"},
		"/products": {Title: "Каталог", Page: "products"},
		"/login":    {Title: "Вход", Page: "login"},
		"/register": {Title: "Регистрация", Page: "register"},
		"/profile":  {Title: "Профиль", Page: "profile"},
		"/cart":     {Title: "Корзина", Page: "cart"},
	}

	for path, config := range pages {
		config.Shop = mTitle
		if path == "/products" || path == "/admin/products" {
			app.Get(path, createProductsHandler(config))
		} else {
			app.Get(path, handlers.CreateDefaultHandler(config))
		}
	}
}

func createProductsHandler(config handlers.PageHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		category := c.Query("category")
		user := c.Locals("user")
		pageNum, _ := strconv.Atoi(c.Query("page", "1"))
		if pageNum < 1 {
			pageNum = 1
		}
		return c.Render(config.Page, fiber.Map{
			"Title":       config.Title + config.Shop,
			"Page":        config.Page,
			"Category":    category,
			"CurrentPage": pageNum,
			"User":        user,
		})
	}
}

func mainInit(apiBaseURL string, apiRoutes []string, title string, port string) *fiber.App {
	port0 := "3000"
	urlDomaine := "http://localhost"
	urlStart := urlDomaine + ":" + port
	webDir := utils.GetWebDir()
	log.Printf("📁 Web directory: %s", webDir)

	app := utils.GetEngineTemplate(webDir, "layouts/main")

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     urlDomaine + ":" + port0 + "," + urlStart,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))
	handler := handlers.NewFrontendHandler(sess)
	app.Use(handler.SessionAuthMiddleware)
	utils.SetupAPIProxy(app, apiBaseURL+"/", apiRoutes)
	utils.SetupStaticFiles(app, webDir)
	log.Println("📁 Using web directory:", webDir)
	setupPageRoutes(app, title)
	// 4. Health check

	// Запуск
	log.Println("🚀 Admin server started on " + urlDomaine + ":" + port)
	log.Fatal(app.Listen(":" + port))
	startApp(app, "admin", title)
	return app
}

func startApp(app *fiber.App, service string, title string) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "healthy",
			"service":   service,
			"timestamp": time.Now(),
		})
	})

	// 5. SPA Fallback — последний

	app.Use(func(c *fiber.Ctx) error {
		path := c.Path()
		if strings.HasPrefix(path, "/api/") ||
			strings.Contains(path, ".") ||
			strings.HasPrefix(path, "/css/") ||
			strings.HasPrefix(path, "/js/") ||
			strings.HasPrefix(path, "/images/") {
			return c.SendStatus(404)
		}
		return c.Render("index", fiber.Map{
			"Title": title,
			"Page":  service,
		})
	})

}
