package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Cluster struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Status    string             `bson:"status" json:"status"`
	IpAddress string             `bson:"ip_address" json:"ip_address"`
	Username  string             `bson:"username" json:"username"`
	Password  string             `bson:"password" json:"password"`
	Port      int                `bson:"port" json:"port"`
	CreatedBy primitive.ObjectID `bson:"created_by" json:"created_by"`
	CreatedAt int64              `bson:"created_at" json:"created_at"`
	UpdatedAt int64              `bson:"updated_at" json:"updated_at"`
}
