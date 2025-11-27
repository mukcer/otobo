package main

import (
	"log"
	"os"
	"otobo/internal/database"
	"otobo/internal/database/repositories"
	"otobo/internal/handlers"
	"otobo/internal/weinkey"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New()
	store := weinkey.ValkeyInit()
	weinkey.SessionInit(store)

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000,http://localhost:3001,http://localhost:3002",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Last-Sync",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// Инициализация базы данных
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Запуск миграций
	if err := db.RunMigrations(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Инициализация репозиториев
	categoryRepo := repositories.NewCategoryRepository(db.DB)
	productRepo := repositories.NewProductRepository(db.DB)
	userRepo := repositories.NewUserRepository(db.DB)
	cartRepo := repositories.NewCartRepository(db.DB)
	orderRepo := repositories.NewOrderRepository(db.DB)
	colorRepo := repositories.NewColorRepository(db.DB)

	// ПРАВИЛЬНАЯ инициализация handlers с dependency injection
	categoryHandler := handlers.NewCategoryHandler(categoryRepo)
	colorHandler := handlers.NewColorHandler(colorRepo)
	productHandler := handlers.NewProductHandler(productRepo, categoryRepo)
	authHandler := handlers.NewAuthHandler(
		userRepo,
		store,
		os.Getenv("JWT_SECRET"),
	)
	cartHandler := handlers.NewCartHandler(cartRepo, productRepo)
	orderHandler := handlers.NewOrderHandler(orderRepo, cartRepo)
	adminHandler := handlers.NewAdminHandler(db, store)

	// Маршруты
	api := app.Group("/api/v1")

	// Админские маршруты (требуют аутентификации и прав администратора)
	admin := api.Group("/admin", authHandler.AuthMiddleware, authHandler.AdminMiddleware)
	admin.Get("/dashboard", adminHandler.GetDashboardStats)
	admin.Get("/database/tables", adminHandler.GetDatabaseTables)
	admin.Get("/database/tables/:tableName/data", adminHandler.GetTableData)
	admin.Put("/database/tables/:tableName/data/:id", adminHandler.UpdateTableData)
	admin.Delete("/database/tables/:tableName", adminHandler.DeleteDatabaseTable)
	admin.Post("/database/backup", adminHandler.BackupDatabase)
	admin.Post("/database/optimize", adminHandler.OptimizeDatabase)
	admin.Post("/database/clear-cache", adminHandler.ClearQueryCache)
	admin.Get("/cache/stats", adminHandler.GetCacheStats)
	admin.Get("/cache/keys", adminHandler.GetCacheKeys)
	admin.Delete("/cache/keys/:key", adminHandler.DeleteCacheKey)
	admin.Post("/cache/clear", adminHandler.ClearAllCache)
	admin.Post("/cache/clear-sessions", adminHandler.ClearSessions)
	// Заказы в админке
	admin.Get("/orders", adminHandler.GetOrders)
	admin.Get("/orders/:id", adminHandler.GetOrderDetails)
	admin.Put("/orders/:id/status", adminHandler.UpdateOrderStatus)

	// Аутентификация
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Товары (публичные)
	products := api.Group("/products")
	products.Get("/categories", categoryHandler.GetCategories)
	products.Get("/", productHandler.GetProducts)
	products.Get("/id/:id", productHandler.GetProductByID)
	products.Get("/:slug", productHandler.GetProduct)

	colors := api.Group("/colors")

	colors.Get("/", colorHandler.GetColors)
	colors.Get("/active", colorHandler.GetActiveColors)
	colors.Get("/:id", colorHandler.GetColorByID)
	colors.Post("/", colorHandler.CreateColor)
	colors.Put("/:id", colorHandler.UpdateColor)
	colors.Delete("/:id", colorHandler.DeleteColor)
	colors.Post("/by-ids", colorHandler.GetColorsByIDs)

	// Корзина (работает для авторизованных и неавторизованных пользователей)
	cart := api.Group("/cart")
	cart.Get("/", cartHandler.GetCart)
	cart.Get("/count", cartHandler.GetCartCount)
	cart.Post("/", cartHandler.AddToCart)
	cart.Put("/:id", cartHandler.UpdateCartItem)
	cart.Delete("/:id", cartHandler.RemoveFromCart)
	cart.Delete("/", cartHandler.ClearCart)

	// Заказы (требует аутентификации)
	orders := api.Group("/orders", handlers.AuthMiddleware)
	orders.Post("/", orderHandler.CreateOrder)
	orders.Get("/", orderHandler.GetUserOrders)
	orders.Get("/:id", orderHandler.GetOrder)
	orders.Put("/:id/status", orderHandler.UpdateOrderStatus)

	// Профиль (требует аутентификации)
	user := api.Group("/user", handlers.AuthMiddleware)
	user.Get("/profile", authHandler.GetProfile)
	user.Get("/sync", authHandler.Sync)
	user.Post("/session", authHandler.CreateSession)
	user.Post("/logout", authHandler.Logout)
	user.Get("/profile", authHandler.GetProfile)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("🛍️  Fashion store server started on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
