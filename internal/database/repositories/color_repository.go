package repositories

import (
	"otobo/internal/models"

	"gorm.io/gorm"
)

type ColorRepository struct {
	db *gorm.DB
}

func NewColorRepository(db *gorm.DB) *ColorRepository {
	return &ColorRepository{db: db}
}

// Create создает новый цвет
func (r *ColorRepository) Create(color *models.Color) error {
	return r.db.Create(color).Error
}

// GetByID возвращает цвет по ID
func (r *ColorRepository) GetByID(id uint) (*models.Color, error) {
	var color models.Color
	err := r.db.First(&color, id).Error
	if err != nil {
		return nil, err
	}
	return &color, nil
}

// GetAll возвращает все цвета
func (r *ColorRepository) GetAll() ([]models.Color, error) {
	var colors []models.Color
	err := r.db.Find(&colors).Error
	if err != nil {
		return nil, err
	}
	return colors, nil
}

// GetActive возвращает только активные цвета (если есть поле статуса)
func (r *ColorRepository) GetActive() ([]models.Color, error) {
	var colors []models.Color
	err := r.db.Where("active = ?", true).Find(&colors).Error
	if err != nil {
		return nil, err
	}
	return colors, nil
}

// Update обновляет цвет
func (r *ColorRepository) Update(color *models.Color) error {
	return r.db.Save(color).Error
}

// Delete удаляет цвет по ID
func (r *ColorRepository) Delete(id uint) error {
	return r.db.Delete(&models.Color{}, id).Error
}

// GetByIDs возвращает цвета по списку ID
func (r *ColorRepository) GetByIDs(ids []uint) ([]models.Color, error) {
	var colors []models.Color
	err := r.db.Where("id IN ?", ids).Find(&colors).Error
	if err != nil {
		return nil, err
	}
	return colors, nil
}

// GetByName возвращает цвет по названию
func (r *ColorRepository) GetByName(name string) (*models.Color, error) {
	var color models.Color
	err := r.db.Where("name = ?", name).First(&color).Error
	if err != nil {
		return nil, err
	}
	return &color, nil
}

// GetByValue возвращает цвет по hex значению
func (r *ColorRepository) GetByValue(value string) (*models.Color, error) {
	var color models.Color
	err := r.db.Where("value = ?", value).First(&color).Error
	if err != nil {
		return nil, err
	}
	return &color, nil
}

// CreateBatch создает несколько цветов сразу
func (r *ColorRepository) CreateBatch(colors []models.Color) error {
	return r.db.Create(&colors).Error
}

// Count возвращает общее количество цветов
func (r *ColorRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Color{}).Count(&count).Error
	return count, err
}

// GetPaginated возвращает цвета с пагинацией
func (r *ColorRepository) GetPaginated(offset, limit int) ([]models.Color, error) {
	var colors []models.Color
	err := r.db.Offset(offset).Limit(limit).Find(&colors).Error
	if err != nil {
		return nil, err
	}
	return colors, nil
}
