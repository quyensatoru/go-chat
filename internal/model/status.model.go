package model

import (
	"time"
)

type Status struct {
	ID        string `json:"id" bson:"_id,omitempty"`
	Name      string `json:"name" bson:"name"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
