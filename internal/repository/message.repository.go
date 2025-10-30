package repository

import (
	"backend/internal/model"
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type MessageRepository struct {
	repo *mongo.Collection
}

func NewMessageRepository(repo *mongo.Collection) *MessageRepository {
	return &MessageRepository{repo: repo}
}

func (r *MessageRepository) FindAll(ctx context.Context) (*[]model.Message, error) {
	var messages []model.Message
	cursor, err := r.repo.Find(ctx, map[string]interface{}{})

	if err != nil {
		return nil, err
	}

	if err = cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return &messages, nil
}

func (r *MessageRepository) Create(ctx context.Context, message *model.Message) (*model.Message, error) {
	_, err := r.repo.InsertOne(ctx, message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (r *MessageRepository) Delete(ctx context.Context, id string) error {
	_, err := r.repo.DeleteOne(ctx, map[string]interface{}{"_id": id})
	if err != nil {
		return err
	}
	return nil
}
