package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Conversation struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	UserIDs   []primitive.ObjectID `bson:"user_ids" json:"user_ids"`
	CreatedAt int64                `bson:"created_at" json:"created_at"`
	UpdatedAt int64                `bson:"updated_at" json:"updated_at"`
}
