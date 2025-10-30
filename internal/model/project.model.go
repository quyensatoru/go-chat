package model

import "time"

type Project struct {
	ID          string `json:"id" bson:"_id,omitempty"`
	Name        string `json:"name" bson:"name"`
	Type        string `json:"type" bson:"type"`
	Access      string `json:"access" bson:"access"`
	Description string `json:"description" bson:"description"`
	TemplateID  string `json:"template" bson:"template"`
	Keyword     string `json:"Keyword" bson:"Keyword"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
