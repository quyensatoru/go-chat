package model

import "time"

type TaskStatus struct {
	ID        string `gorm:"primaryKey" json:"id" bson:"_id,omitempty"`
	TaskID    string `json:"task_id" bson:"task_id"`
	StatusID  string `gorm:"primaryKey" json:"status_id" bson:"status_id"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
