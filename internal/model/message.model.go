package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Content     string             `json:"content" bson:"content"`
	SenderID    primitive.ObjectID `json:"sender_id" bson:"sender_id"`
	RecepientID primitive.ObjectID `json:"recepient_id" bson:"recepient_id"`
	TaskID      string             `json:"task_id" bson:"task_id"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
