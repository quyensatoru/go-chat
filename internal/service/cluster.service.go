package service

import (
	"backend/internal/repository"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ClusterService struct {
	repository *repository.ClusterRepository
}

func NewClusterService(repo *repository.ClusterRepository) *ClusterService {
	return &ClusterService{repository: repo}
}

func (s *ClusterService) CreateCluster(context context.Context, doc bson.M) (*mongo.InsertOneResult, error) {
	result, err := s.repository.Create(context, doc, nil)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *ClusterService) FindOne(ctx context.Context, filter bson.M) (bson.M, error) {
	result, err := s.repository.FindOne(ctx, filter)
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

func (s *ClusterService) FindAll(ctx context.Context, filter bson.M) ([]bson.M, error) {
	cursor, err := s.repository.FindAll(ctx, filter, nil)
	if err != nil {
		return nil, err
	}
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}
