package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Development struct {
	ID        string             `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	ClusterID primitive.ObjectID `bson:"cluster_id" json:"cluster_id"`
	Status    string             `bson:"status" json:"status"`
	CreatedAt int64              `bson:"created_at" json:"created_at"`
	UpdatedAt int64              `bson:"updated_at" json:"updated_at"`
}
