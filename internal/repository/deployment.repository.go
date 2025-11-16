package repository

import (
	"backend/internal/model"
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DeploymentRepository struct {
	model *mongo.Collection
}

func NewDeploymentRepository(collection *mongo.Collection) *DeploymentRepository {
	return &DeploymentRepository{model: collection}
}

func (repo *DeploymentRepository) Create(ctx context.Context, deployment interface{}) (*mongo.InsertOneResult, error) {
	return repo.model.InsertOne(ctx, deployment)
}

func (repo *DeploymentRepository) FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult {
	return repo.model.FindOne(ctx, filter)
}

func (repo *DeploymentRepository) FindAll(ctx context.Context, filter interface{}, option *options.FindOptions) *[]model.Development {
	cursor, err := repo.model.Find(ctx, filter, option)

	if err != nil {
		return nil
	}

	var results []model.Development

	if err = cursor.All(ctx, &results); err != nil {
		return nil
	}
	return &results
}

func (repo *DeploymentRepository) CreateMany(ctx context.Context, deployments []interface{}) (*mongo.InsertManyResult, error) {
	return repo.model.InsertMany(ctx, deployments)
}
