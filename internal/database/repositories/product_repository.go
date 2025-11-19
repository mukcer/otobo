package repositories

import (
	"fmt"

	"otobo/internal/models"

	"gorm.io/gorm"
)

type ProductRepository struct {
	DB *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{DB: db}
}

type ProductFilter struct {
	CategorySlug string
	Size         string
	Color        string
	MinPrice     float64
	MaxPrice     float64
	InStock      bool
	IsActive     bool
}

func (r *ProductRepository) FindAll(filter ProductFilter, page, limit int, sortBy, order string) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	query := r.DB.Model(&models.Product{})

	// Применяем фильтры
	if filter.CategorySlug != "" {
		query = query.Joins("JOIN categories ON categories.id = products.category_id").
			Where("categories.slug = ?", filter.CategorySlug)
	}

	if filter.Size != "" {
		query = query.Joins("JOIN product_sizes ON product_sizes.product_id = products.id").
			Joins("JOIN sizes ON sizes.id = product_sizes.size_id").
			Where("sizes.value = ?", filter.Size)
	}

	if filter.Color != "" {
		query = query.Joins("JOIN product_colors ON product_colors.product_id = products.id").
			Joins("JOIN colors ON colors.id = product_colors.color_id").
			Where("colors.value = ?", filter.Color)
	}

	if filter.MinPrice > 0 {
		query = query.Where("price >= ?", filter.MinPrice)
	}

	if filter.MaxPrice > 0 {
		query = query.Where("price <= ?", filter.MaxPrice)
	}

	if filter.InStock {
		query = query.Where("in_stock = ?", true)
	}

	if filter.IsActive {
		query = query.Where("is_active = ?", true)
	}

	// Считаем общее количество
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Применяем сортировку и пагинацию
	offset := (page - 1) * limit
	if err := query.
		Preload("Category").
		Preload("Sizes").
		Preload("Colors").
		Order(fmt.Sprintf("%s %s", sortBy, order)).
		Offset(offset).
		Limit(limit).
		Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *ProductRepository) FindBySlug(slug string) (*models.Product, error) {
	var product models.Product

	err := r.DB.
		Preload("Category").
		Preload("Sizes").
		Preload("Colors").
		Preload("Variations").
		Preload("Variations.Size").
		Preload("Variations.Color").
		Where("slug = ?", slug).
		First(&product).Error

	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *ProductRepository) Create(product *models.Product) error {
	return r.DB.Create(product).Error
}

func (r *ProductRepository) Update(product *models.Product) error {
	return r.DB.Save(product).Error
}

func (r *ProductRepository) Delete(id uint) error {
	return r.DB.Delete(&models.Product{}, id).Error
}

// internal/database/repositories/product_repository.go (дополнение)
func (r *ProductRepository) FindByCategory(categoryID uint) ([]models.Product, error) {
	var products []models.Product
	err := r.DB.Preload("Category").Preload("Sizes").Preload("Colors").
		Where("category_id = ? AND is_active = ?", categoryID, true).
		Order("created_at DESC").
		Find(&products).Error
	return products, err
}

func (r *ProductRepository) FindFeatured(limit int) ([]models.Product, error) {
	var products []models.Product
	err := r.DB.Preload("Category").Preload("Sizes").Preload("Colors").
		Where("is_active = ? AND in_stock = ?", true, true).
		Order("created_at DESC").
		Limit(limit).
		Find(&products).Error
	return products, err
}

func (r *ProductRepository) UpdateStock(variationID uint, quantity int) error {
	return r.DB.Model(&models.ProductVariation{}).
		Where("id = ?", variationID).
		Update("quantity", gorm.Expr("quantity + ?", quantity)).Error
}

func (r *ProductRepository) Search(query string, page, limit int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	searchQuery := r.DB.Preload("Category").Preload("Sizes").Preload("Colors").
		Where("is_active = ?", true)

	if query != "" {
		searchQuery = searchQuery.Where(
			"name ILIKE ? OR description ILIKE ? OR sku ILIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%",
		)
	}

	// Считаем общее количество
	if err := searchQuery.Model(&models.Product{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := searchQuery.Offset(offset).Limit(limit).Find(&products).Error

	return products, total, err
}
