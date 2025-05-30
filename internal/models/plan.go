package models

import (
	"time"

	"gorm.io/datatypes"
)

type Plan struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"size:100;not null" json:"name"`
	Price         float64        `gorm:"not null" json:"price"`
	Features      datatypes.JSON `gorm:"type:jsonb" json:"features" swaggertype:"object"`
	Duration      int            `gorm:"column:duration_days" json:"duration_days"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	Subscriptions []Subscription `json:"-"`
}
