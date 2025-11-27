package repositories

import (
	"errors"
	"time"

	"otobo/internal/models"

	"gorm.io/gorm"
)

type CartRepository struct {
	DB *gorm.DB
}

func NewCartRepository(db *gorm.DB) *CartRepository {
	return &CartRepository{DB: db}
}

// GetCart возвращает корзину пользователя или создает новую
func (r *CartRepository) GetCart(userID uint, sessionID string) (*models.Cart, error) {
	var cart models.Cart

	query := r.DB.Preload("Items").
		Preload("Items.Product").
		Preload("Items.ProductVariation").
		Preload("Items.ProductVariation.Size").
		Preload("Items.ProductVariation.Color").
		Preload("Items.ProductVariation.Product") // Добавляем загрузку Product

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	} else {
		query = query.Where("session_id = ?", sessionID)
	}

	if err := query.First(&cart).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.createCart(userID, sessionID)
		}
		return nil, err
	}

	return &cart, nil
}

// createCart создает новую корзину
func (r *CartRepository) createCart(userID uint, sessionID string) (*models.Cart, error) {
	cart := &models.Cart{
		UserID:    userID,
		SessionID: sessionID,
	}

	if err := r.DB.Create(cart).Error; err != nil {
		return nil, err
	}

	return cart, nil
}

// AddItem добавляет товар в корзину
func (r *CartRepository) AddItem(userID uint, sessionID string, productID, variationID uint, quantity int) (*models.Cart, error) {
	// Получаем или создаем корзину
	cart, err := r.GetCart(userID, sessionID)
	if err != nil {
		return nil, err
	}

	// Проверяем наличие вариации товара и загружаем связанный продукт
	var variation models.ProductVariation
	if err := r.DB.
		Preload("Product"). // Загружаем связанный продукт
		Preload("Size").
		Preload("Color").
		First(&variation, variationID).Error; err != nil {
		return nil, errors.New("product variation not found")
	}

	// Проверяем, что связанный товар активен и в наличии
	if !variation.Product.IsActive {
		return nil, errors.New("product is not active")
	}

	if !variation.Product.InStock {
		return nil, errors.New("product is out of stock")
	}

	// Проверяем доступное количество конкретной вариации
	if variation.Quantity < quantity {
		return nil, errors.New("not enough stock for this variation")
	}

	// Проверяем, есть ли уже такой товар в корзине
	var existingItem models.CartItem
	err = r.DB.Where("cart_id = ? AND product_variation_id = ?", cart.ID, variationID).First(&existingItem).Error

	if err == nil {
		// Обновляем количество существующего товара
		newQuantity := existingItem.Quantity + quantity
		if variation.Quantity < newQuantity {
			return nil, errors.New("not enough stock for additional quantity")
		}

		existingItem.Quantity = newQuantity
		if err := r.DB.Save(&existingItem).Error; err != nil {
			return nil, err
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Добавляем новый товар в корзину
		cartItem := models.CartItem{
			CartID:             cart.ID,
			ProductID:          productID,
			ProductVariationID: variationID,
			Quantity:           quantity,
		}

		if err := r.DB.Create(&cartItem).Error; err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	// Обновляем время корзины
	cart.UpdatedAt = time.Now()
	r.DB.Save(cart)

	// Возвращаем обновленную корзину
	return r.GetCart(userID, sessionID)
}

// UpdateItem обновляет количество товара в корзине
func (r *CartRepository) UpdateItem(cartID, itemID uint, quantity int) error {
	if quantity <= 0 {
		return r.RemoveItem(itemID)
	}

	var item models.CartItem
	if err := r.DB.
		Preload("ProductVariation").
		Preload("ProductVariation.Product"). // Загружаем продукт через вариацию
		First(&item, itemID).Error; err != nil {
		return errors.New("cart item not found")
	}

	// Проверяем доступное количество
	if item.ProductVariation.Quantity < quantity {
		return errors.New("not enough stock")
	}

	// Проверяем, что товар все еще активен
	if !item.ProductVariation.Product.IsActive {
		return errors.New("product is no longer available")
	}

	item.Quantity = quantity
	return r.DB.Save(&item).Error
}

// RemoveItem удаляет товар из корзины
func (r *CartRepository) RemoveItem(itemID uint) error {
	return r.DB.Delete(&models.CartItem{}, itemID).Error
}

// ClearCart очищает корзину
func (r *CartRepository) ClearCart(cartID uint) error {
	return r.DB.Where("cart_id = ?", cartID).Delete(&models.CartItem{}).Error
}

// GetCartItemsCount возвращает общее количество товаров в корзине
func (r *CartRepository) GetCartItemsCount(userID uint, sessionID string) (int64, error) {
	cart, err := r.GetCart(userID, sessionID)
	if err != nil {
		return 0, err
	}

	var count int64
	for _, item := range cart.Items {
		count += int64(item.Quantity)
	}

	return count, nil
}

// CalculateCartTotal вычисляет общую стоимость корзины
func (r *CartRepository) CalculateCartTotal(userID uint, sessionID string) (float64, error) {
	cart, err := r.GetCart(userID, sessionID)
	if err != nil {
		return 0, err
	}

	var total float64
	for _, item := range cart.Items {
		// Используем цену из Product, который теперь загружается через ProductVariation
		if item.ProductVariation.Product.ID > 0 {
			total += item.ProductVariation.Product.Price * float64(item.Quantity)
		}
	}

	return total, nil
}

// MergeCarts объединяет корзины (при входе пользователя)
func (r *CartRepository) MergeCarts(userID uint, sessionID string) error {
	// Находим корзину по session_id
	var sessionCart models.Cart
	if err := r.DB.Preload("Items").Where("session_id = ?", sessionID).First(&sessionCart).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // Нет корзины для объединения
		}
		return err
	}

	// Находим корзину пользователя
	var userCart models.Cart
	if err := r.DB.Preload("Items").Where("user_id = ?", userID).First(&userCart).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Просто обновляем session cart с user_id
			sessionCart.UserID = userID
			sessionCart.SessionID = ""
			return r.DB.Save(&sessionCart).Error
		}
		return err
	}

	// Объединяем товары из session cart в user cart
	for _, sessionItem := range sessionCart.Items {
		var existingItem models.CartItem
		err := r.DB.Where("cart_id = ? AND product_variation_id = ?", userCart.ID, sessionItem.ProductVariationID).
			First(&existingItem).Error

		if err == nil {
			// Товар уже есть в корзине пользователя - обновляем количество
			existingItem.Quantity += sessionItem.Quantity

			// Проверяем доступное количество
			var variation models.ProductVariation
			if err := r.DB.Preload("Product").First(&variation, sessionItem.ProductVariationID).Error; err != nil {
				continue // Пропускаем если вариация не найдена
			}

			if variation.Quantity >= existingItem.Quantity {
				if err := r.DB.Save(&existingItem).Error; err != nil {
					return err
				}
			}
			// Удаляем item из session cart
			r.DB.Delete(&sessionItem)
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			// Товара нет в корзине пользователя - перемещаем
			sessionItem.CartID = userCart.ID
			if err := r.DB.Save(&sessionItem).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Удаляем session cart
	return r.DB.Delete(&sessionCart).Error
}

// ValidateCartItems проверяет все товары в корзине на доступность
func (r *CartRepository) ValidateCartItems(userID uint, sessionID string) ([]string, error) {
	cart, err := r.GetCart(userID, sessionID)
	if err != nil {
		return nil, err
	}

	var warnings []string

	for _, item := range cart.Items {
		var variation models.ProductVariation
		if err := r.DB.
			Preload("Product").
			Preload("Size").
			Preload("Color").
			First(&variation, item.ProductVariationID).Error; err != nil {

			warnings = append(warnings, "Product variation not found")
			continue
		}

		if !variation.Product.IsActive {
			warnings = append(warnings, "Product is no longer available")
		}

		if !variation.Product.InStock {
			warnings = append(warnings, "Product is out of stock")
		}

		if variation.Quantity < item.Quantity {
			warnings = append(warnings, "Not enough stock for product")
		}
	}

	return warnings, nil
}

// GetByID возвращает корзину по ID
func (r *CartRepository) GetByID(cartID uint) (*models.Cart, error) {
	var cart models.Cart
	if err := r.DB.Preload("Items").
		Preload("Items.Product").
		Preload("Items.ProductVariation").
		Preload("Items.ProductVariation.Size").
		Preload("Items.ProductVariation.Color").
		First(&cart, cartID).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}
