package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ConversationRepository struct {
	collection *mongo.Collection
}

func NewConversationRepository(collection *mongo.Collection) *ConversationRepository {
	return &ConversationRepository{collection: collection}
}

func (repo *ConversationRepository) FindOneAndUpdate(ctx context.Context, filter bson.M, update bson.M, opts *options.FindOneAndUpdateOptions) (*mongo.SingleResult, error) {
	singleResult := repo.collection.FindOneAndUpdate(ctx, filter, update, opts)
	return singleResult, nil
}
