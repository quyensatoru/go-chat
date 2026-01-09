package repository

import (
	"backend/internal/model"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type NotificationRepository interface {
	Create(notification *model.Notification) error
	FindByUserID(userID uint) (model.Notification, error)
	FindByID(id uint) (*model.Notification, error)
	FindAll() ([]model.Notification, error)
	Update(notification *model.Notification) error
	CreateOrUpdate(notification *model.Notification) error
	Delete(id uint) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(notification *model.Notification) error {
	return r.db.Create(notification).Error
}

func (r *notificationRepository) CreateOrUpdate(notification *model.Notification) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"p256dh": notification.P256dh, "auth": notification.Auth, "endpoint": notification.Endpoint}),
	}).Create(notification).Error
}

func (r *notificationRepository) FindByUserID(userID uint) (model.Notification, error) {
	var notification model.Notification
	err := r.db.Where("user_id = ?", userID).First(&notification).Error
	return notification, err
}

func (r *notificationRepository) FindByID(id uint) (*model.Notification, error) {
	var notification model.Notification
	err := r.db.First(&notification, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &notification, nil
}

func (r *notificationRepository) FindAll() ([]model.Notification, error) {
	var notifications []model.Notification
	err := r.db.Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) Update(notification *model.Notification) error {
	return r.db.Save(notification).Error
}

func (r *notificationRepository) Delete(id uint) error {
	return r.db.Delete(&model.Notification{}, id).Error
}
