package model

import "time"

type Task struct {
	ID            string `json:"id" bson:"_id,omitempty"`
	Description   string `json:"description" bson:"description"`
	ProjectID     string `json:"project_id" bson:"project_id"`
	Priority      string `json:"priority" bson:"priority"`
	Status        string `json:"status" bson:"status"`
	Label         string `json:"label" bson:"label"`
	TaskIndex     int    `json:"task_index" bson:"task_index"`
	Link          string `json:"link" bson:"link"`
	Archived      bool   `json:"archived" bson:"archived"`
	Deleted       bool   `json:"deleted" bson:"deleted"`
	Assignee      string `json:"assignee" bson:"assignee"`
	Reporter      string `json:"reporter" bson:"reporter"`
	StartDate     string `json:"start_date" bson:"start_date"`
	DueDate       string `json:"due_date" bson:"due_date"`
	DevelopmentID string `json:"development_id" bson:"development_id"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
