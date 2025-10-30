package repository

import (
	"backend/internal/model"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository struct {
	Collection *mongo.Collection
}

func NewUserRepository(collection *mongo.Collection) *UserRepository {
	return &UserRepository{Collection: collection}
}

func (doc *UserRepository) GetAll(ctx context.Context, filter map[string]interface{}) ([]*model.User, error) {
	var users []*model.User
	cursor, err := doc.Collection.Find(ctx, filter)

	if err != nil {
		return nil, err
	}

	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (doc *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	var user = &model.User{}

	objectId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}

	err = doc.Collection.FindOne(ctx, map[string]any{"_id": objectId}).Decode(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (doc *UserRepository) FindOne(ctx context.Context, filter bson.M) (*model.User, error) {
	var user = &model.User{}
	err := doc.Collection.FindOne(ctx, filter).Decode(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (doc *UserRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	_, err := doc.Collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (doc *UserRepository) Update(ctx context.Context, id string, user *model.User) (*model.User, error) {
	_, err := doc.Collection.UpdateOne(ctx, map[string]interface{}{"_id": id}, map[string]interface{}{"$set": user})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (doc *UserRepository) Delete(ctx context.Context, id string) error {
	_, err := doc.Collection.DeleteOne(ctx, map[string]interface{}{"_id": id})
	if err != nil {
		return err
	}
	return nil
}

func (doc *UserRepository) FindOneAndUpdate(ctx context.Context, filter bson.M, update bson.M, options *options.FindOneAndUpdateOptions) (*model.User, error) {
	var result model.User

	err := doc.Collection.FindOneAndUpdate(ctx, filter, update, options).Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}
