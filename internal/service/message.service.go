package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"
)

type MessageService struct {
	svc *repository.MessageRepository
}

func NewMessageService(svc *repository.MessageRepository) *MessageService {
	return &MessageService{svc: svc}
}

func (s *MessageService) GetAll(ctx context.Context) (*[]model.Message, error) {
	return s.svc.FindAll(ctx)
}

func (s *MessageService) Create(ctx context.Context, message *model.Message) (*model.Message, error) {
	return s.svc.Create(ctx, message)
}
