package model

import (
	"time"
)

type Server struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"type:varchar(255);not null"`
	Status    string    `json:"status" gorm:"type:varchar(50);default:'inactive'"`
	IpAddress string    `json:"ip_address" gorm:"type:varchar(45);not null"`
	Username  string    `json:"username" gorm:"type:varchar(100);not null"`
	Password  string    `json:"password" gorm:"type:varchar(255);not null"`
	Port      int       `json:"port" gorm:"not null;default:22"`
	CreatedBy uint      `json:"created_by" gorm:"not null;index"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	// Relationships
	Creator User  `json:"creator,omitempty" gorm:"foreignKey:CreatedBy;constraint:OnDelete:CASCADE"`
	Apps    []App `json:"apps,omitempty" gorm:"foreignKey:ServerID;constraint:OnDelete:CASCADE"`
}
