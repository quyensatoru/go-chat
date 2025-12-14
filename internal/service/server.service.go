package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"errors"
)

type ServerService interface {
	CreateServer(server *model.Server) error
	GetServerByID(id uint) (*model.Server, error)
	GetServersByCreatedBy(createdBy uint) ([]model.Server, error)
	GetAllServers() ([]model.Server, error)
	UpdateServer(server *model.Server) error
	DeleteServer(id uint) error
	ValidateServerAccess(serverID, userID uint) error
}

type serverService struct {
	repo repository.ServerRepository
}

func NewServerService(repo repository.ServerRepository) ServerService {
	return &serverService{repo: repo}
}

func (s *serverService) CreateServer(server *model.Server) error {
	// Set default status if not provided
	if server.Status == "" {
		server.Status = "inactive"
	}
	return s.repo.Create(server)
}

func (s *serverService) GetServerByID(id uint) (*model.Server, error) {
	return s.repo.FindByID(id)
}

func (s *serverService) GetServersByCreatedBy(createdBy uint) ([]model.Server, error) {
	return s.repo.FindByCreatedBy(createdBy)
}

func (s *serverService) GetAllServers() ([]model.Server, error) {
	return s.repo.FindAll()
}

func (s *serverService) UpdateServer(server *model.Server) error {
	return s.repo.Update(server)
}

func (s *serverService) DeleteServer(id uint) error {
	return s.repo.Delete(id)
}

func (s *serverService) ValidateServerAccess(serverID, userID uint) error {
	server, err := s.repo.FindByID(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		return errors.New("server not found")
	}
	if server.CreatedBy != userID {
		return errors.New("access denied: user is not the creator of this server")
	}
	return nil
}
