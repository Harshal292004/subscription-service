package models

import "time"

type User struct {
	ID           uint         `gorm:"primaryKey" json:"id"`
	Name         string       `gorm:"size:100;not null" json:"name"`
	Password     string       `gorm:"not null" json:"-"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	Subscription Subscription `gorm:"foreignKey:UserID" json:"subscription"`
}
