package repositories

import (
	"otobo/internal/models"

	"gorm.io/gorm"
)

type CategoryRepository struct {
	DB *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{DB: db}
}

func (r *CategoryRepository) FindAll() ([]models.Category, error) {
	var categories []models.Category
	err := r.DB.Order("name ASC").Find(&categories).Error
	return categories, err
}

func (r *CategoryRepository) FindBySlug(slug string) (*models.Category, error) {
	var category models.Category
	err := r.DB.Where("slug = ?", slug).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepository) FindWithProducts() ([]models.Category, error) {
	var categories []models.Category
	err := r.DB.Preload("Products", "is_active = ?", true).
		Where("EXISTS (SELECT 1 FROM products WHERE products.category_id = categories.id AND products.is_active = true)").
		Order("name ASC").
		Find(&categories).Error
	return categories, err
}

func (r *CategoryRepository) Create(category *models.Category) error {
	return r.DB.Create(category).Error
}

func (r *CategoryRepository) Update(category *models.Category) error {
	return r.DB.Save(category).Error
}

func (r *CategoryRepository) Delete(id uint) error {
	// Проверяем, есть ли товары в категории
	var count int64
	r.DB.Model(&models.Product{}).Where("category_id = ?", id).Count(&count)

	if count > 0 {
		// return errors.New("cannot delete category with existing products")
	}

	return r.DB.Delete(&models.Category{}, id).Error
}
