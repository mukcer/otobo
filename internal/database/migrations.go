// internal/database/migrations.go
package database

import (
	"fmt"
	"log"

	"otobo/internal/models"

	"gorm.io/gorm"
)

// RunMigrations выполняет все миграции
func (db *Database) RunMigrations() error {
	log.Println("Running database migrations...")

	err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Size{},
		&models.Color{},
		&models.Product{},
		&models.ProductVariation{},
		&models.Cart{},
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.Review{},
	)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("✅ Database migrations completed")
	return nil
}

// SeedDatabase заполняет базу начальными данными
func (db *Database) SeedDatabase() error {
	log.Println("Seeding database with initial data...")

	// Запускаем в транзакции
	return db.WithTransaction(func(tx *gorm.DB) error {
		if err := db.seedSizes(tx); err != nil {
			return err
		}
		if err := db.seedColors(tx); err != nil {
			return err
		}
		if err := db.seedCategories(tx); err != nil {
			return err
		}
		if err := db.seedAdminUser(tx); err != nil {
			return err
		}

		log.Println("✅ Database seeding completed")
		return nil
	})
}

func (db *Database) seedSizes(tx *gorm.DB) error {
	sizes := []models.Size{
		{Name: "Extra Small", Value: "xs"},
		{Name: "Small", Value: "s"},
		{Name: "Medium", Value: "m"},
		{Name: "Large", Value: "l"},
		{Name: "Extra Large", Value: "xl"},
		{Name: "XX Large", Value: "xxl"},
	}

	for _, size := range sizes {
		var existing models.Size
		if err := tx.Where("value = ?", size.Value).First(&existing).Error; err != nil {
			if err := tx.Create(&size).Error; err != nil {
				return err
			}
		}
	}

	log.Println("✅ Sizes seeded")
	return nil
}

func (db *Database) seedColors(tx *gorm.DB) error {
	colors := []models.Color{
		{Name: "Черный", Value: "#000000"},
		{Name: "Белый", Value: "#FFFFFF"},
		{Name: "Красный", Value: "#FF0000"},
		{Name: "Синий", Value: "#0000FF"},
		{Name: "Зеленый", Value: "#008000"},
		{Name: "Розовый", Value: "#FFC0CB"},
		{Name: "Бежевый", Value: "#F5F5DC"},
		{Name: "Серый", Value: "#808080"},
	}

	for _, color := range colors {
		var existing models.Color
		if err := tx.Where("value = ?", color.Value).First(&existing).Error; err != nil {
			if err := tx.Create(&color).Error; err != nil {
				return err
			}
		}
	}

	log.Println("✅ Colors seeded")
	return nil
}

func (db *Database) seedCategories(tx *gorm.DB) error {
	categories := []models.Category{
		{
			Name:        "Платья",
			Slug:        "dresses",
			Description: "Элегантные платья для любого случая",
			ImageURL:    "/images/categories/dresses.jpg",
		},
		{
			Name:        "Блузки и рубашки",
			Slug:        "blouses",
			Description: "Стильные блузки и рубашки",
			ImageURL:    "/images/categories/blouses.jpg",
		},
		{
			Name:        "Юбки",
			Slug:        "skirts",
			Description: "Женственные юбки разных фасонов",
			ImageURL:    "/images/categories/skirts.jpg",
		},
		{
			Name:        "Брюки",
			Slug:        "pants",
			Description: "Удобные и модные брюки",
			ImageURL:    "/images/categories/pants.jpg",
		},
		{
			Name:        "Верхняя одежда",
			Slug:        "outerwear",
			Description: "Пальто, куртки и пуховики",
			ImageURL:    "/images/categories/outerwear.jpg",
		},
		{
			Name:        "Аксессуары",
			Slug:        "accessories",
			Description: "Сумки, ремни и украшения",
			ImageURL:    "/images/categories/accessories.jpg",
		},
	}

	for _, category := range categories {
		var existing models.Category
		if err := tx.Where("slug = ?", category.Slug).First(&existing).Error; err != nil {
			if err := tx.Create(&category).Error; err != nil {
				return err
			}
		}
	}

	log.Println("✅ Categories seeded")
	return nil
}

func (db *Database) seedAdminUser(tx *gorm.DB) error {
	// В реальном приложении пароль должен хэшироваться!
	adminUser := models.User{
		Email:     "admin@fashionstore.com",
		Password:  "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
		FirstName: "Admin",
		LastName:  "FashionStore",
		Role:      "admin",
	}

	var existing models.User
	if err := tx.Where("email = ?", adminUser.Email).First(&existing).Error; err != nil {
		if err := tx.Create(&adminUser).Error; err != nil {
			return err
		}
	}

	log.Println("✅ Admin user seeded")
	return nil
}
