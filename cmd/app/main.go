package main

import (
	"log"
	"os"

	"otobo/internal/database"
	"otobo/internal/database/repositories"
	"otobo/internal/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/redis/go-redis/v9"
)

func main() {
	app := fiber.New()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", //localhost
		Password: "",
		DB:       0,
	})
	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000,http://localhost:3001",
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

	// ПРАВИЛЬНАЯ инициализация handlers с dependency injection
	categoryHandler := handlers.NewCategoryHandler(categoryRepo)
	productHandler := handlers.NewProductHandler(productRepo, categoryRepo)
	authHandler := handlers.NewAuthHandler(
		userRepo,
		rdb,
		os.Getenv("JWT_SECRET"),
	)
	cartHandler := handlers.NewCartHandler(cartRepo, productRepo)
	orderHandler := handlers.NewOrderHandler(orderRepo, cartRepo)

	// Маршруты
	api := app.Group("/api/v1")

	// Аутентификация
	auth := api.Group("/auth", authHandler.AuthMiddleware)
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Товары (публичные)
	products := api.Group("/products")
	products.Get("/categories", categoryHandler.GetCategories)
	products.Get("/", productHandler.GetProducts)
	products.Get("/id/:id", productHandler.GetProductByID)
	products.Get("/:slug", productHandler.GetProduct)

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
