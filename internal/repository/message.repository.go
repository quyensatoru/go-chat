package repository

import (
	"backend/internal/model"
	"errors"

	"gorm.io/gorm"
)

type MessageRepository interface {
	Create(message *model.Message) error
	FindByID(id uint) (*model.Message, error)
	FindBySenderID(senderID uint) ([]model.Message, error)
	FindByRecipientID(recipientID uint) ([]model.Message, error)
	FindByChannel(channelID string, limit int, offset int) ([]model.MessageView, error)
	FindAll() ([]model.Message, error)
	Update(message *model.Message) error
	Delete(id uint) error
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(message *model.Message) error {
	return r.db.Create(message).Error
}

func (r *messageRepository) FindByID(id uint) (*model.Message, error) {
	var message model.Message
	err := r.db.Preload("Sender").Preload("Recipient").First(&message, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &message, nil
}

func (r *messageRepository) FindBySenderID(senderID uint) ([]model.Message, error) {
	var messages []model.Message
	err := r.db.Where("sender_id = ?", senderID).Preload("Sender").Preload("Recipient").Find(&messages).Error
	return messages, err
}

func (r *messageRepository) FindByRecipientID(recipientID uint) ([]model.Message, error) {
	var messages []model.Message
	err := r.db.Where("recipient_id = ?", recipientID).Preload("Sender").Preload("Recipient").Find(&messages).Error
	return messages, err
}

func (r *messageRepository) FindByChannel(channelID string, limit int, offset int) ([]model.MessageView, error) {
	var messages []model.MessageView
	err := r.db.
		Model(&model.Message{}).
		Select(`
			id, content, recipient_id AS target_id, sender_id, created_at
		`).
		Where("channel_id = ?", channelID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).
		Error
	return messages, err
}

func (r *messageRepository) FindAll() ([]model.Message, error) {
	var messages []model.Message
	err := r.db.Preload("Sender").Preload("Recipient").Find(&messages).Error
	return messages, err
}

func (r *messageRepository) Update(message *model.Message) error {
	return r.db.Save(message).Error
}

func (r *messageRepository) Delete(id uint) error {
	return r.db.Delete(&model.Message{}, id).Error
}
