package repository

import (
	"backend/internal/model"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EnviromentRepository struct {
	model *mongo.Collection
}

func NewEnviromentRepository(model *mongo.Collection) *EnviromentRepository {
	return &EnviromentRepository{model: model}
}

func (repo *EnviromentRepository) Find(ctx context.Context, filter bson.M, options *options.FindOptions) *[]model.Enviroment {
	cursor, err := repo.model.Find(ctx, filter, options)

	if err != nil {
		return nil
	}
	var enviroments []model.Enviroment
	if err = cursor.All(ctx, &enviroments); err != nil {
		return nil
	}
	return &enviroments
}

func (repo *EnviromentRepository) Create(ctx context.Context, doc bson.M, options *options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return repo.model.InsertOne(ctx, doc, options)
}

func (repo *EnviromentRepository) FindOneAndUpdate(ctx context.Context, filter bson.M, update bson.M, option *options.FindOneAndUpdateOptions) (*mongo.SingleResult, error) {
	singleResult := repo.model.FindOneAndUpdate(ctx, filter, update, option)
	return singleResult, nil
}
