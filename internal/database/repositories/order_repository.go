package repositories

import (
	"errors"
	"fmt"
	"time"

	"otobo/internal/models"

	"gorm.io/gorm"
)

type OrderRepository struct {
	DB *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{DB: db}
}

func (r *OrderRepository) CreateFromCart(userID uint, shippingAddress, shippingMethod, paymentMethod string) (*models.Order, error) {
	// Получаем корзину пользователя
	var cart models.Cart
	if err := r.DB.
		Preload("Items").
		Preload("Items.ProductVariation").
		Preload("Items.ProductVariation.Product"). // Добавляем загрузку продукта
		Preload("Items.ProductVariation.Size").
		Preload("Items.ProductVariation.Color").
		Where("user_id = ?", userID).First(&cart).Error; err != nil {
		return nil, errors.New("cart not found")
	}

	if len(cart.Items) == 0 {
		return nil, errors.New("cart is empty")
	}

	// Считаем общую сумму и проверяем доступность
	var totalAmount float64
	var orderItems []models.OrderItem

	for _, cartItem := range cart.Items {
		// Используем цену из продукта через вариацию
		productPrice := cartItem.ProductVariation.Product.Price
		itemTotal := productPrice * float64(cartItem.Quantity)
		totalAmount += itemTotal

		orderItem := models.OrderItem{
			ProductID:          cartItem.ProductID,
			ProductVariationID: cartItem.ProductVariationID,
			Quantity:           cartItem.Quantity,
			Price:              productPrice, // Используем цену продукта
		}
		orderItems = append(orderItems, orderItem)

		// Обновляем количество товара на складе
		if err := r.updateProductStock(cartItem.ProductVariationID, cartItem.Quantity); err != nil {
			return nil, err
		}
	}
	// Рассчитываем стоимость доставки
	shippingCost := r.calculateShippingCost(totalAmount, shippingMethod)
	taxAmount := totalAmount * 0.2 // 20% НДС (пример)
	finalAmount := totalAmount + shippingCost + taxAmount

	// Создаем заказ
	order := &models.Order{
		UserID:          userID,
		OrderNumber:     r.generateOrderNumber(),
		Status:          "pending",
		TotalAmount:     totalAmount,
		ShippingCost:    shippingCost,
		TaxAmount:       taxAmount,
		Discount:        0,
		FinalAmount:     finalAmount,
		ShippingAddress: shippingAddress,
		ShippingMethod:  shippingMethod,
		PaymentMethod:   paymentMethod,
		PaymentStatus:   "pending",
		OrderItems:      orderItems,
	}

	// Создаем заказ в транзакции
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// Очищаем корзину
		if err := tx.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return r.FindByID(order.ID)
}

func (r *OrderRepository) updateProductStock(variationID uint, quantity int) error {
	var variation models.ProductVariation
	if err := r.DB.First(&variation, variationID).Error; err != nil {
		return err
	}

	if variation.Quantity < quantity {
		return errors.New("not enough stock")
	}

	variation.Quantity -= quantity
	return r.DB.Save(&variation).Error
}

func (r *OrderRepository) calculateShippingCost(totalAmount float64, shippingMethod string) float64 {
	switch shippingMethod {
	case "express":
		return 500.0
	case "standard":
		if totalAmount > 5000 {
			return 0 // Бесплатная доставка от 5000 руб
		}
		return 300.0
	default:
		return 300.0
	}
}

func (r *OrderRepository) generateOrderNumber() string {
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("ORD-%s", timestamp)
}

func (r *OrderRepository) FindByID(orderID uint) (*models.Order, error) {
	var order models.Order
	err := r.DB.Preload("User").Preload("OrderItems").Preload("OrderItems.Product").
		Preload("OrderItems.ProductVariation").Preload("OrderItems.ProductVariation.Size").
		Preload("OrderItems.ProductVariation.Color").
		First(&order, orderID).Error

	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *OrderRepository) FindByUserID(userID uint) ([]models.Order, error) {
	var orders []models.Order
	err := r.DB.Preload("OrderItems").Preload("OrderItems.Product").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&orders).Error

	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *OrderRepository) UpdateStatus(orderID uint, status string) error {
	return r.DB.Model(&models.Order{}).Where("id = ?", orderID).Update("status", status).Error
}

func (r *OrderRepository) UpdatePaymentStatus(orderID uint, paymentStatus string) error {
	return r.DB.Model(&models.Order{}).Where("id = ?", orderID).Update("payment_status", paymentStatus).Error
}

func (r *OrderRepository) FindAll(filter OrderFilter) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.DB.Preload("User").Preload("OrderItems")

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.UserID > 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}

	if !filter.StartDate.IsZero() {
		query = query.Where("created_at >= ?", filter.StartDate)
	}

	if !filter.EndDate.IsZero() {
		query = query.Where("created_at <= ?", filter.EndDate)
	}

	// Считаем общее количество
	if err := query.Model(&models.Order{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Применяем пагинацию
	offset := (filter.Page - 1) * filter.Limit
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(filter.Limit).
		Find(&orders).Error

	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

type OrderFilter struct {
	Status    string
	UserID    uint
	StartDate time.Time
	EndDate   time.Time
	Page      int
	Limit     int
}

// GetOrderByID возвращает заказ по ID
func (r *OrderRepository) GetOrderByID(orderID uint) (*models.Order, error) {
	var order models.Order
	err := r.DB.Preload("User").
		Preload("OrderItems").
		Preload("OrderItems.Product").
		Preload("OrderItems.ProductVariation").
		Preload("OrderItems.ProductVariation.Size").
		Preload("OrderItems.ProductVariation.Color").
		First(&order, orderID).Error

	if err != nil {
		return nil, err
	}

	return &order, nil
}

// UpdateOrderStatus обновляет статус заказа
func (r *OrderRepository) UpdateOrderStatus(orderID uint, status string) error {
	return r.DB.Model(&models.Order{}).Where("id = ?", orderID).Update("status", status).Error
}

// UpdateOrderPaymentStatus обновляет статус оплаты заказа
func (r *OrderRepository) UpdateOrderPaymentStatus(orderID uint, paymentStatus string) error {
	return r.DB.Model(&models.Order{}).Where("id = ?", orderID).Update("payment_status", paymentStatus).Error
}
