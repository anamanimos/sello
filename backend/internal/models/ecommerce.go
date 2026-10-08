package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CouponType enum
type CouponType string

const (
	CouponTypePercentage CouponType = "percentage"
	CouponTypeFixed      CouponType = "fixed"
)

// Coupon represents a discount coupon
type Coupon struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	StoreID     *uuid.UUID `gorm:"type:uuid;index" json:"store_id"` // Null if Platform Coupon (Subscription)
	Code        string     `gorm:"size:50;not null;index" json:"code"`
	Type        CouponType `gorm:"size:20;default:'fixed'" json:"type"`
	Value       float64    `gorm:"type:decimal(10,2);not null" json:"value"`
	MinPurchase float64    `gorm:"type:decimal(10,2);default:0" json:"min_purchase"`
	MaxDiscount float64    `gorm:"type:decimal(10,2);default:0" json:"max_discount"` // For percentage
	UsageLimit  int        `gorm:"default:0" json:"usage_limit"`                     // 0 = unlimited
	UsedCount   int        `gorm:"default:0" json:"used_count"`
	ExpiresAt   *time.Time `json:"expires_at"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Coupon) TableName() string {
	return "coupons"
}

// OrderStatus enum
type OrderStatus string

const (
	OrderStatusPending             OrderStatus = "pending"
	OrderStatusWaitingConfirmation OrderStatus = "waiting_confirmation"
	OrderStatusPaid                OrderStatus = "paid"
	OrderStatusShipped             OrderStatus = "shipped"
	OrderStatusCompleted           OrderStatus = "completed"
	OrderStatusCancelled           OrderStatus = "cancelled"
    OrderStatusExpired             OrderStatus = "expired"
)

// Order represents a transaction
type Order struct {
	ID            uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	StoreID       uuid.UUID   `gorm:"type:uuid;not null;index" json:"store_id"`
	Store         Store       `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	
	// Customer Info (Guest or User)
	CustomerName  string      `gorm:"size:100;not null" json:"customer_name"`
	CustomerPhone string      `gorm:"size:20;not null" json:"customer_phone"`
	CustomerEmail string      `gorm:"size:100" json:"customer_email"`
	
	// Payment
	TotalAmount    float64     `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	DiscountAmount float64     `gorm:"type:decimal(10,2);default:0" json:"discount_amount"`
	FinalAmount    float64     `gorm:"type:decimal(10,2);not null" json:"final_amount"`
	Status         OrderStatus `gorm:"size:20;default:'pending';index" json:"status"`
	PaymentMethod  string      `gorm:"size:50" json:"payment_method"`
	PaymentProof   string      `gorm:"type:text" json:"payment_proof"` // URL to image
	
	// Coupon
	CouponCode string `gorm:"size:50" json:"coupon_code"`
	CouponID   *uint  `json:"coupon_id"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	// Relations
	Items []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}

// OrderItem represents items in an order
type OrderItem struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderID     uuid.UUID `gorm:"type:uuid;not null;index" json:"order_id"`
	ProductID   uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	ProductName string    `gorm:"size:200;not null" json:"product_name"` // Snapshot name
	VariantName string    `gorm:"size:100" json:"variant_name"`
	Price       float64   `gorm:"type:decimal(10,2);not null" json:"price"` // Snapshot price
	Quantity    int       `gorm:"not null" json:"quantity"`
	Subtotal    float64   `gorm:"type:decimal(10,2);not null" json:"subtotal"`
}

func (OrderItem) TableName() string {
	return "order_items"
}

// Input Structs
type CreateOrderItemInput struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
	Variant   string `json:"variant"`
}

type CreateOrderInput struct {
	CustomerName  string                 `json:"customer_name" binding:"required"`
	CustomerPhone string                 `json:"customer_phone" binding:"required"`
	CustomerEmail string                 `json:"customer_email"`
	Items         []CreateOrderItemInput `json:"items" binding:"required,dive"`
	CouponCode    string                 `json:"coupon_code"`
	PaymentMethod string                 `json:"payment_method" binding:"required"`
}

type CreateCouponInput struct {
	Code        string     `json:"code" binding:"required"`
	Type        CouponType `json:"type" binding:"required"`
	Value       float64    `json:"value" binding:"required"`
	MinPurchase float64    `json:"min_purchase"`
	MaxDiscount float64    `json:"max_discount"`
	UsageLimit  int        `json:"usage_limit"`
	ExpiresAt   string     `json:"expires_at"` // RFC3339 or YYYY-MM-DD
}
