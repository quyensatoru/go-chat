package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Enviroment struct {
	ID           string             `bson:"_id,omitempty" json:"id"`
	DataSet      string             `bson:"dataset" json:"dataset"` // Object storage enviroment data
	DeploymentID primitive.ObjectID `bson:"deployment_id" json:"deployment_id"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}
