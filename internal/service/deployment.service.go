package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type DeploymentService struct {
	repo *repository.DeploymentRepository
}

func NewDeploymentService(repo *repository.DeploymentRepository) *DeploymentService {
	return &DeploymentService{repo: repo}
}

func (s *DeploymentService) Create(ctx context.Context, deployment bson.M) (*mongo.InsertOneResult, error) {
	return s.repo.Create(ctx, deployment)
}

func (s *DeploymentService) FindOne(ctx context.Context, filter bson.M) *mongo.SingleResult {
	return s.repo.FindOne(ctx, filter)
}

func (s *DeploymentService) FindAll(ctx context.Context, filter bson.M) *[]model.Development {
	return s.repo.FindAll(ctx, filter, nil)
}

func (s *DeploymentService) CreateMany(ctx context.Context, deployments []interface{}) (*mongo.InsertManyResult, error) {
	return s.repo.CreateMany(ctx, deployments)
}
