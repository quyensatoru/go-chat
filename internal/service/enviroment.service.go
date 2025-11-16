package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EnviromentService struct {
	repo *repository.EnviromentRepository
}

func NewEnviromentService(repo *repository.EnviromentRepository) *EnviromentService {
	return &EnviromentService{repo: repo}
}

func (s *EnviromentService) Create(ctx context.Context, doc bson.M) (*mongo.InsertOneResult, error) {
	return s.repo.Create(ctx, doc, nil)
}

func (s *EnviromentService) Find(ctx context.Context, filter bson.M) (*[]model.Enviroment, error) {
	enviroments := s.repo.Find(ctx, filter, nil)
	if enviroments == nil {
		return nil, mongo.ErrNoDocuments
	}

	return enviroments, nil
}

func (s *EnviromentService) FindOneAndUpdate(ctx context.Context, filter bson.M, update bson.M) (*model.Enviroment, error) {
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	singleResult, err := s.repo.FindOneAndUpdate(ctx, filter, update, opts)
	if err != nil {
		return nil, err
	}
	var enviroment model.Enviroment
	err = singleResult.Decode(&enviroment)
	if err != nil {
		return nil, err
	}
	return &enviroment, nil
}
