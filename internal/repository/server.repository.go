package repository

import (
	"backend/internal/model"
	"errors"

	"gorm.io/gorm"
)

type ServerRepository interface {
	Create(server *model.Server) error
	FindByID(id uint) (*model.Server, error)
	FindByCreatedBy(createdBy uint) ([]model.Server, error)
	FindAll() ([]model.Server, error)
	Update(server *model.Server) error
	Delete(id uint) error
}

type serverRepository struct {
	db *gorm.DB
}

func NewServerRepository(db *gorm.DB) ServerRepository {
	return &serverRepository{db: db}
}

func (r *serverRepository) Create(server *model.Server) error {
	return r.db.Create(server).Error
}

func (r *serverRepository) FindByID(id uint) (*model.Server, error) {
	var server model.Server
	err := r.db.Preload("Creator").Preload("Apps").First(&server, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &server, nil
}

func (r *serverRepository) FindByCreatedBy(createdBy uint) ([]model.Server, error) {
	var servers []model.Server
	err := r.db.Where("created_by = ?", createdBy).Preload("Creator").Preload("Apps").Find(&servers).Error
	return servers, err
}

func (r *serverRepository) FindAll() ([]model.Server, error) {
	var servers []model.Server
	err := r.db.Preload("Creator").Preload("Apps").Find(&servers).Error
	return servers, err
}

func (r *serverRepository) Update(server *model.Server) error {
	return r.db.Save(server).Error
}

func (r *serverRepository) Delete(id uint) error {
	return r.db.Delete(&model.Server{}, id).Error
}
