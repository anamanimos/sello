package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubscriptionStatus string

const (
	SubscriptionPending   SubscriptionStatus = "pending"
	SubscriptionPaid      SubscriptionStatus = "paid"
	SubscriptionActive    SubscriptionStatus = "active" // Approved
	SubscriptionRejected  SubscriptionStatus = "rejected"
	SubscriptionExpired   SubscriptionStatus = "expired"
)

// SubscriptionTransaction represents a merchant purchasing a plan
type SubscriptionTransaction struct {
	ID            uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID          `gorm:"type:uuid;not null;index" json:"user_id"`
	User          User               `gorm:"foreignKey:UserID" json:"user"`
	PlanID        uint               `gorm:"not null" json:"plan_id"`
	Plan          Plan               `gorm:"foreignKey:PlanID" json:"plan"`
	Amount        float64            `gorm:"type:decimal(10,2);not null" json:"amount"`
	DurationMonth int                `gorm:"default:1" json:"duration_month"` // 1 or 12
	Status        SubscriptionStatus `gorm:"size:20;default:'pending';index" json:"status"`
	PaymentMethod string             `gorm:"size:50" json:"payment_method"`
	PaymentProof  string             `gorm:"type:text" json:"payment_proof"`
    
    // Coupon
    CouponID *uint   `json:"coupon_id"`
    Coupon   *Coupon `gorm:"foreignKey:CouponID" json:"coupon,omitempty"`
	
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	DeletedAt     gorm.DeletedAt     `gorm:"index" json:"-"`
}

func (SubscriptionTransaction) TableName() string {
	return "subscription_transactions"
}
