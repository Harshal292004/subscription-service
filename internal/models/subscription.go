package models

import "time"

type SubscriptionStatus string

const (
	Active    SubscriptionStatus = "ACTIVE"
	Inactive  SubscriptionStatus = "INACTIVE"
	Cancelled SubscriptionStatus = "CANCELLED"
	Expired   SubscriptionStatus = "EXPIRED"
)

type Subscription struct {
	ID        uint               `gorm:"primaryKey" json:"id"`
	UserID    uint               `gorm:"not null;unique" json:"user_id"`
	PlanID    uint               `gorm:"not null" json:"plan_id"`
	Status    SubscriptionStatus `gorm:"type:subscription_status;not null"`
	StartDate time.Time          `gorm:"not null" json:"start_date"`
	EndDate   time.Time          `gorm:"not null" json:"end_date"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
	User      *User              `gorm:"foreignKey:UserID" json:"-"`
	Plan      *Plan              `gorm:"foreignKey:PlanID" json:"-"`
}
