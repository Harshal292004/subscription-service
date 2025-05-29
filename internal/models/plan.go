package models

import "time"

type Plan struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `json:"name"`
	Price         float64        `json:"price"`
	Features      []string       `gorm:"type:text[]" json:"features"`
	Duration      int            `json:"duration_days"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	Subscriptions []Subscription `json:"-"`
}
