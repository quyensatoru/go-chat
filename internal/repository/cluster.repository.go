package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ClusterRepository struct {
	collection *mongo.Collection
}

func NewClusterRepository(collection *mongo.Collection) *ClusterRepository {
	return &ClusterRepository{collection: collection}
}

func (c *ClusterRepository) Create(context context.Context, doc bson.M, options *options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return c.collection.InsertOne(context, doc, options)
}

func (c *ClusterRepository) FindOne(ctx context.Context, filter bson.M) (*mongo.SingleResult, error) {
	singleResult := c.collection.FindOne(ctx, filter)
	return singleResult, nil
}

func (c *ClusterRepository) FindAll(ctx context.Context, filter bson.M, opts *options.FindOptions) (*mongo.Cursor, error) {
	cursor, err := c.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	return cursor, nil
}
