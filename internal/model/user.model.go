package model

import (
	"time"
)

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"type:varchar(100);not null;uniqueIndex"`
	Email     string    `json:"email" gorm:"type:varchar(255);not null;uniqueIndex"`
	Password  string    `json:"-" gorm:"type:varchar(255);not null"`
	Token     string    `json:"token" gorm:"type:text"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}
