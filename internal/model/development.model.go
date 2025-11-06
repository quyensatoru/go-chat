package model

type Development struct {
	ID        string `bson:"_id,omitempty" json:"id"`
	Name      string `bson:"name" json:"name"`
	PodID     string `bson:"pod_id" json:"pod_id"`
	ClusterID string `bson:"cluster_id" json:"cluster_id"`
	Status    string `bson:"status" json:"status"`
	CreatedAt int64  `bson:"created_at" json:"created_at"`
	UpdatedAt int64  `bson:"updated_at" json:"updated_at"`
}
