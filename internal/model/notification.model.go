package model

import (
	"time"
)

type Notification struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Endpoint  string    `json:"endpoint" gorm:"type:text;not null"`
	P256dh    string    `json:"p256dh" gorm:"type:text;not null"`
	Auth      string    `json:"auth" gorm:"type:text;not null"`
	UserID    uint      `json:"user_id" gorm:"not null;index;unique"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}
