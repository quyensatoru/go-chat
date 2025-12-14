package repository

import (
	"backend/internal/model"
	"errors"

	"gorm.io/gorm"
)

type AppRepository interface {
	Create(app *model.App) error
	FindByID(id uint) (*model.App, error)
	FindByServerID(serverID uint) ([]model.App, error)
	FindAll() ([]model.App, error)
	Update(app *model.App) error
	Delete(id uint) error
}

type appRepository struct {
	db *gorm.DB
}

func NewAppRepository(db *gorm.DB) AppRepository {
	return &appRepository{db: db}
}

func (r *appRepository) Create(app *model.App) error {
	return r.db.Create(app).Error
}

func (r *appRepository) FindByID(id uint) (*model.App, error) {
	var app model.App
	err := r.db.Preload("Server").First(&app, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &app, nil
}

func (r *appRepository) FindByServerID(serverID uint) ([]model.App, error) {
	var apps []model.App
	err := r.db.Where("server_id = ?", serverID).Preload("Server").Find(&apps).Error
	return apps, err
}

func (r *appRepository) FindAll() ([]model.App, error) {
	var apps []model.App
	err := r.db.Preload("Server").Find(&apps).Error
	return apps, err
}

func (r *appRepository) Update(app *model.App) error {
	return r.db.Save(app).Error
}

func (r *appRepository) Delete(id uint) error {
	return r.db.Delete(&model.App{}, id).Error
}
