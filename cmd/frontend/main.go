package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

func main() {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,http://localhost:3001",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// Получаем путь к статическим файлам фронтенда
	webDir := getWebDir()

	// Статические файлы
	app.Use("/", filesystem.New(filesystem.Config{
		Root:   http.Dir(webDir),
		Index:  "index.html",
		MaxAge: 3600,
	}))

	// Прокси для API запросов к основному серверу
	setupAPIProxy(app)

	// Специальные маршруты для SPA (Single Page Application)
	setupSPARoutes(app, webDir)

	log.Println("🚀 Frontend server started on :3001")
	log.Printf("📁 Serving static files from: %s", webDir)
	log.Fatal(app.Listen(":3001"))
}

func getWebDir() string {
	// Пытаемся найти папку web относительно текущей директории
	dirsToCheck := []string{
		"./web",
		"../web",
		"../../web",
		"./cmd/frontend/web",
	}

	for _, dir := range dirsToCheck {
		if _, err := os.Stat(dir); err == nil {
			absPath, _ := filepath.Abs(dir)
			return absPath
		}
	}

	// Если папка не найдена, создаем базовую структуру
	log.Println("⚠️  Web directory not found, creating basic structure...")
	return createBasicWebStructure()
}

func createBasicWebStructure() string {
	baseDir := "./web"
	os.MkdirAll(baseDir, 0755)

	// Создаем базовые папки
	dirs := []string{"css", "js", "images"}
	for _, dir := range dirs {
		os.MkdirAll(filepath.Join(baseDir, dir), 0755)
	}

	// Создаем базовый index.html
	createBasicIndexHTML(baseDir)
	createBasicCSS(baseDir)
	createBasicJS(baseDir)

	absPath, _ := filepath.Abs(baseDir)
	return absPath
}

func setupAPIProxy(app *fiber.App) {
	// Прокси для API endpoints
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
			// Основной сервер на порту 3000
			targetURL := "http://localhost:3000" + c.Path()

			if err := proxy.Do(c, targetURL); err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": "API server unavailable",
				})
			}

			// Remove Server header from response
			c.Response().Header.Del(fiber.HeaderServer)
			return nil
		})
	}

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "healthy",
			"service":   "frontend",
			"timestamp": time.Now(),
		})
	})
}

func setupSPARoutes(app *fiber.App, webDir string) {
	// Маршруты для SPA - все неизвестные пути возвращают index.html
	spaRoutes := []string{
		"/login",
		"/register",
		"/profile",
		"/products",
		"/cart",
		"/admin",
	}

	for _, route := range spaRoutes {
		app.Get(route, func(c *fiber.Ctx) error {
			return c.SendFile(filepath.Join(webDir, "index.html"))
		})
	}
}

func createBasicIndexHTML(baseDir string) {
	htmlContent := `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Fashion Store - Магазин модной женской одежды</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap" rel="stylesheet">
    <link rel="stylesheet" href="/css/style.css">
</head>
<body>
    <nav class="navbar">
        <div class="container">
            <div class="navbar-brand">
                <a href="/" class="logo">
                    <i class="fas fa-crown"></i>
                    FashionStore
                </a>
            </div>
            <div class="navbar-menu">
                <a href="/products" class="nav-link">
                    <i class="fas fa-shopping-bag"></i>
                    Магазин
                </a>
                <div class="auth-links">
                    <a href="/login" class="nav-link">
                        <i class="fas fa-sign-in-alt"></i>
                        Войти
                    </a>
                    <a href="/register" class="nav-link register-btn">
                        Регистрация
                    </a>
                </div>
            </div>
        </div>
    </nav>

    <main class="main-content">
        <section class="hero">
            <div class="container">
                <div class="hero-content">
                    <h1>Добро пожаловать в Fashion Store</h1>
                    <p>Сервер фронтенда запущен успешно! Файлы будут обслуживаться из папки web/</p>
                    <div class="hero-buttons">
                        <a href="/products" class="btn btn-primary">
                            <i class="fas fa-shopping-bag"></i>
                            Перейти в магазин
                        </a>
                        <a href="/login" class="btn btn-secondary">
                            <i class="fas fa-sign-in-alt"></i>
                            Войти в систему
                        </a>
                    </div>
                </div>
            </div>
        </section>
    </main>

    <footer class="footer">
        <div class="container">
            <p>&copy; 2024 Fashion Store. Frontend served by Go server.</p>
        </div>
    </footer>

    <script src="/js/main.js"></script>
</body>
</html>`

	os.WriteFile(filepath.Join(baseDir, "index.html"), []byte(htmlContent), 0644)
}

func createBasicCSS(baseDir string) {
	cssContent := `* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: 'Inter', sans-serif;
    line-height: 1.6;
    color: #333;
    background-color: #f8f9fa;
}

.container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 1rem;
}

/* Navbar */
.navbar {
    background: white;
    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
    padding: 1rem 0;
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    z-index: 1000;
}

.navbar .container {
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.logo {
    font-size: 1.5rem;
    font-weight: 700;
    color: #e91e63;
    text-decoration: none;
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.navbar-menu {
    display: flex;
    align-items: center;
    gap: 2rem;
}

.nav-link {
    color: #333;
    text-decoration: none;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 1rem;
    border-radius: 8px;
    transition: all 0.3s ease;
}

.nav-link:hover {
    background: #f8f9fa;
    color: #e91e63;
}

.register-btn {
    background: #e91e63;
    color: white !important;
}

.register-btn:hover {
    background: #d81b60;
}

/* Hero Section */
.hero {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    padding: 8rem 0 4rem;
    text-align: center;
    min-height: 100vh;
    display: flex;
    align-items: center;
}

.hero-content h1 {
    font-size: 3rem;
    margin-bottom: 1rem;
    font-weight: 700;
}

.hero-content p {
    font-size: 1.2rem;
    margin-bottom: 2rem;
    opacity: 0.9;
}

.hero-buttons {
    display: flex;
    gap: 1rem;
    justify-content: center;
    flex-wrap: wrap;
}

.btn {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 2rem;
    text-decoration: none;
    border-radius: 8px;
    font-weight: 600;
    transition: all 0.3s ease;
}

.btn-primary {
    background: #e91e63;
    color: white;
}

.btn-primary:hover {
    background: #d81b60;
    transform: translateY(-2px);
}

.btn-secondary {
    background: transparent;
    color: white;
    border: 2px solid white;
}

.btn-secondary:hover {
    background: white;
    color: #333;
}

/* Footer */
.footer {
    background: #333;
    color: white;
    text-align: center;
    padding: 2rem 0;
}

/* Responsive */
@media (max-width: 768px) {
    .hero-content h1 {
        font-size: 2rem;
    }
    
    .hero-buttons {
        flex-direction: column;
        align-items: center;
    }
    
    .btn {
        width: 200px;
        justify-content: center;
    }
}`

	os.WriteFile(filepath.Join(baseDir, "css", "style.css"), []byte(cssContent), 0644)
}

func createBasicJS(baseDir string) {
	jsContent := `console.log('Fashion Store frontend loaded successfully');

// Basic navigation
document.addEventListener('DOMContentLoaded', function() {
    console.log('DOM fully loaded and parsed');
    
    // Add loading states to buttons
    const buttons = document.querySelectorAll('.btn, .nav-link');
    buttons.forEach(button => {
        button.addEventListener('click', function(e) {
            if (this.href && !this.href.startsWith('http')) {
                console.log('Navigating to:', this.href);
                // Add loading state here if needed
            }
        });
    });
});

// Basic error handling
window.addEventListener('error', function(e) {
    console.error('JavaScript error:', e.error);
});

// API health check
async function checkAPIHealth() {
    try {
        const response = await fetch('/health');
        const data = await response.json();
        console.log('Frontend server health:', data);
    } catch (error) {
        console.error('Health check failed:', error);
    }
}

// Run health check on load
checkAPIHealth();`

	os.WriteFile(filepath.Join(baseDir, "js", "main.js"), []byte(jsContent), 0644)
}
