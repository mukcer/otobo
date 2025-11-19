package models

import (
	"time"

	"gorm.io/gorm"
)

// Пользователь
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Email     string         `json:"email" gorm:"uniqueIndex"`
	Password  string         `json:"-"`
	FirstName string         `json:"first_name"`
	LastName  string         `json:"last_name"`
	Phone     string         `json:"phone"`
	Address   string         `json:"address"`
	Role      string         `json:"role" gorm:"default:customer"` // customer, admin
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Связи
	Orders  []Order  `json:"orders,omitempty" gorm:"foreignKey:UserID"`
	Reviews []Review `json:"reviews,omitempty" gorm:"foreignKey:UserID"`
}

// Категория товара
type Category struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug" gorm:"uniqueIndex"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`

	Products []Product `json:"products,omitempty" gorm:"foreignKey:CategoryID"`
}

// Размеры одежды
type Size struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Name  string `json:"name"`  // XS, S, M, L, XL
	Value string `json:"value"` // xs, s, m, l, xl
}

// Цвета
type Color struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Name  string `json:"name"`  // Черный, Белый, Красный
	Value string `json:"value"` // #000000, #FFFFFF
}

// Товар
type Product struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug" gorm:"uniqueIndex"`
	Description  string    `json:"description"`
	Price        float64   `json:"price"`
	ComparePrice float64   `json:"compare_price"` // Цена до скидки
	SKU          string    `json:"sku" gorm:"uniqueIndex"`
	Barcode      string    `json:"barcode"`
	CategoryID   uint      `json:"category_id"`
	Category     Category  `json:"category" gorm:"foreignKey:CategoryID"`
	InStock      bool      `json:"in_stock" gorm:"default:true"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	Images       []string  `json:"images" gorm:"type:json"`
	Features     []string  `json:"features" gorm:"type:json"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Many-to-many связи
	Sizes      []Size             `json:"sizes" gorm:"many2many:product_sizes;"`
	Colors     []Color            `json:"colors" gorm:"many2many:product_colors;"`
	Variations []ProductVariation `json:"variations,omitempty" gorm:"foreignKey:ProductID"`
}

// Вариация товара (размер + цвет + количество)
type ProductVariation struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	ProductID uint   `json:"product_id"`
	SizeID    uint   `json:"size_id"`
	ColorID   uint   `json:"color_id"`
	Quantity  int    `json:"quantity"`
	ImageURL  string `json:"image_url"`

	// Связи
	Product Product `json:"product" gorm:"foreignKey:ProductID"`
	Size    Size    `json:"size" gorm:"foreignKey:SizeID"`
	Color   Color   `json:"color" gorm:"foreignKey:ColorID"`
}

// Корзина
type Cart struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	SessionID string    `json:"session_id"` // Для неавторизованных пользователей
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Items []CartItem `json:"items,omitempty" gorm:"foreignKey:CartID"`
}

// Элемент корзины
type CartItem struct {
	ID                 uint `json:"id" gorm:"primaryKey"`
	CartID             uint `json:"cart_id"`
	ProductID          uint `json:"product_id"`
	ProductVariationID uint `json:"product_variation_id"`
	Quantity           int  `json:"quantity"`

	Product          Product          `json:"product" gorm:"foreignKey:ProductID"`
	ProductVariation ProductVariation `json:"variation" gorm:"foreignKey:ProductVariationID"`
}

// Заказ
type Order struct {
	ID           uint    `json:"id" gorm:"primaryKey"`
	UserID       uint    `json:"user_id"`
	OrderNumber  string  `json:"order_number" gorm:"uniqueIndex"`
	Status       string  `json:"status"` // pending, paid, shipped, delivered, cancelled
	TotalAmount  float64 `json:"total_amount"`
	ShippingCost float64 `json:"shipping_cost"`
	TaxAmount    float64 `json:"tax_amount"`
	Discount     float64 `json:"discount"`
	FinalAmount  float64 `json:"final_amount"`

	// Информация о доставке
	ShippingAddress string `json:"shipping_address"`
	ShippingMethod  string `json:"shipping_method"`

	// Информация об оплате
	PaymentMethod string `json:"payment_method"`
	PaymentStatus string `json:"payment_status"` // pending, paid, failed

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User       User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
	OrderItems []OrderItem `json:"items,omitempty" gorm:"foreignKey:OrderID"`
}

// Элемент заказа
type OrderItem struct {
	ID                 uint    `json:"id" gorm:"primaryKey"`
	OrderID            uint    `json:"order_id"`
	ProductID          uint    `json:"product_id"`
	ProductVariationID uint    `json:"product_variation_id"`
	Quantity           int     `json:"quantity"`
	Price              float64 `json:"price"`

	Product          Product          `json:"product" gorm:"foreignKey:ProductID"`
	ProductVariation ProductVariation `json:"variation" gorm:"foreignKey:ProductVariationID"`
}

// Отзыв
type Review struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	ProductID uint      `json:"product_id"`
	Rating    int       `json:"rating"` // 1-5
	Comment   string    `json:"comment"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`

	User    User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Product Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}
