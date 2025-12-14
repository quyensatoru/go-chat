package model

import (
	"time"
)

type Message struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Content     string    `json:"content" gorm:"type:text;not null"`
	SenderID    uint      `json:"sender_id" gorm:"not null;index"`
	RecipientID uint      `json:"recipient_id" gorm:"not null;index"`
	TaskID      string    `json:"task_id" gorm:"type:varchar(255)"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
	
	// Relationships
	Sender    User `json:"sender,omitempty" gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE"`
	Recipient User `json:"recipient,omitempty" gorm:"foreignKey:RecipientID;constraint:OnDelete:CASCADE"`
}
