package model

type Pod struct {
	ID           string `bson:"_id,omitempty" json:"id"`
	Name         string `bson:"name" json:"name"`
	Namespace    string `bson:"namespace" json:"namespace"`
	ClusterID    string `bson:"cluster_id" json:"cluster_id"`
	DeploymentID string `bson:"deployment_id" json:"deployment_id"`
	Status       string `bson:"status" json:"status"`
	CreatedAt    int64  `bson:"created_at" json:"created_at"`
	UpdatedAt    int64  `bson:"updated_at" json:"updated_at"`
}
