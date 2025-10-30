package service

import (
	"backend/internal/repository"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ConversationService struct {
	repo *repository.ConversationRepository
}

func NewConversationService(repo *repository.ConversationRepository) *ConversationService {
	return &ConversationService{repo: repo}
}

func (s *ConversationService) FindOneAndUpdate(ctx context.Context, filter bson.M, update bson.M, opts *options.FindOneAndUpdateOptions) (bson.M, error) {
	result, err := s.repo.FindOneAndUpdate(ctx, filter, update, opts)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	err = result.Decode(&data)

	if err != nil {
		return nil, err
	}

	return data, nil
}
