package model

import (
	"time"
)

type ServiceConfig struct {
	Name   string `json:"name"`
	EnvRaw string `json:"env_raw" gorm:"serializer:json"` // Raw environment variables as string
}

type App struct {
	ID        uint            `json:"id" gorm:"primaryKey"`
	Name      string          `json:"name" gorm:"type:varchar(255);not null"`
	HelmChart string          `json:"helm_chart" gorm:"type:varchar(500)"`
	ServerID  uint            `json:"server_id" gorm:"not null;index"`
	Status    string          `json:"status" gorm:"type:varchar(50);default:'inactive'"`
	Services  []ServiceConfig `json:"services" gorm:"serializer:json"`
	CreatedAt time.Time       `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`

	// Relationships
	Server Server `json:"server,omitempty" gorm:"foreignKey:ServerID;constraint:OnDelete:CASCADE"`
}
