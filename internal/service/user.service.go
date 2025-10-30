package service

import (
	contextkey "backend/internal/common/contextKey"
	"backend/internal/model"
	"backend/internal/repository"
	"context"
	"fmt"
	"time"

	"firebase.google.com/go/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) GetAll(ctx context.Context) ([]*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)

	defer cancel()

	token, ok := ctx.Value(contextkey.UserFirebase).(*auth.Token)

	if !ok {
		return nil, fmt.Errorf("unauthorized: missing user UID in context")
	}

	email := token.Claims["email"].(string)

	filter := bson.M{
		"email": bson.M{
			"$ne": email,
		},
	}

	users, err := s.repo.GetAll(ctx, filter)

	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *UserService) GetByID(ctx context.Context, id string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return s.repo.FindOne(ctx, bson.M{
		"email": email,
	})
}

func (s *UserService) CreateByFirbase(ctx context.Context, payload *auth.Token) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"email":     payload.Claims["email"].(string),
			"username":  payload.Claims["name"].(string),
			"token":     payload.UID,
			"updatedAt": time.Now(),
		},
	}

	filter := bson.M{"email": payload.Claims["email"].(string)}

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	user, err := s.repo.FindOneAndUpdate(ctx, filter, update, opts)

	if err != nil {
		return nil, fmt.Errorf("failed upsert new account %v", err)
	}

	return user, nil
}

func (s *UserService) Update(ctx context.Context, id string, user *model.User) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return s.repo.Update(ctx, id, user)
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return s.repo.Delete(ctx, id)
}
