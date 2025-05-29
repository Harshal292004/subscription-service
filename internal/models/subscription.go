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
	UserID    uint               `gorm:"uniqueIndex" json:"user_id"`
	PlanID    uint               `json:"plan_id"`
	Status    SubscriptionStatus `gorm:"type:varchar(20)" json:"status"`
	StartDate time.Time          `json:"start_date"`
	EndDate   time.Time          `json:"end_date"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
	User      *User              `gorm:"foreignKey:UserID" json:"-"`
	Plan      *Plan              `gorm:"foreignKey:PlanID" json:"-"`
}
