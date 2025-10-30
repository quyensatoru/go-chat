package service

import (
	"context"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
)

type FirebaseService struct {
	App  *firebase.App
	Auth *auth.Client
}

func NewFirebaseService(app *firebase.App) (*FirebaseService, error) {
	auth, error := app.Auth(context.Background())

	if error != nil {
		return nil, error
	}

	return &FirebaseService{
		App:  app,
		Auth: auth,
	}, nil
}

func (fb *FirebaseService) VerifyToken(ctx context.Context, token string) (*auth.Token, error) {
	verify, err := fb.Auth.VerifyIDToken(ctx, token)

	if err != nil {
		return nil, err
	}

	return verify, nil
}

func (fb *FirebaseService) GetUser(ctx context.Context, uid string) (*auth.UserRecord, error) {
	user, err := fb.Auth.GetUser(ctx, uid)

	if err != nil {
		return nil, err
	}

	return user, nil
}
