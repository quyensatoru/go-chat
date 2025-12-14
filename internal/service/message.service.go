package service

import (
	"backend/internal/model"
	"backend/internal/repository"
)

type MessageService interface {
	CreateMessage(message *model.Message) error
	GetMessageByID(id uint) (*model.Message, error)
	GetMessagesBySenderID(senderID uint) ([]model.Message, error)
	GetMessagesByRecipientID(recipientID uint) ([]model.Message, error)
	GetMessagesByTaskID(taskID string) ([]model.Message, error)
	GetAllMessages() ([]model.Message, error)
	UpdateMessage(message *model.Message) error
	DeleteMessage(id uint) error
}

type messageService struct {
	repo repository.MessageRepository
}

func NewMessageService(repo repository.MessageRepository) MessageService {
	return &messageService{repo: repo}
}

func (s *messageService) CreateMessage(message *model.Message) error {
	return s.repo.Create(message)
}

func (s *messageService) GetMessageByID(id uint) (*model.Message, error) {
	return s.repo.FindByID(id)
}

func (s *messageService) GetMessagesBySenderID(senderID uint) ([]model.Message, error) {
	return s.repo.FindBySenderID(senderID)
}

func (s *messageService) GetMessagesByRecipientID(recipientID uint) ([]model.Message, error) {
	return s.repo.FindByRecipientID(recipientID)
}

func (s *messageService) GetMessagesByTaskID(taskID string) ([]model.Message, error) {
	return s.repo.FindByTaskID(taskID)
}

func (s *messageService) GetAllMessages() ([]model.Message, error) {
	return s.repo.FindAll()
}

func (s *messageService) UpdateMessage(message *model.Message) error {
	return s.repo.Update(message)
}

func (s *messageService) DeleteMessage(id uint) error {
	return s.repo.Delete(id)
}
