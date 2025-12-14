package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"errors"
	"fmt"
	"log"
)

type AppService interface {
	CreateApp(app *model.App) error
	GetAppByID(id uint) (*model.App, error)
	GetAppsByServerID(serverID uint) ([]model.App, error)
	GetAllApps() ([]model.App, error)
	UpdateApp(app *model.App) error
	DeleteApp(id uint) error
	ValidateAppServerExists(serverID uint) error
}

type appService struct {
	repo              repository.AppRepository
	serverRepo        repository.ServerRepository
	automationService ServerAutomationService
}

func NewAppService(repo repository.AppRepository, serverRepo repository.ServerRepository, automationService ServerAutomationService) AppService {
	return &appService{
		repo:              repo,
		serverRepo:        serverRepo,
		automationService: automationService,
	}
}

func (s *appService) CreateApp(app *model.App) error {
	// Validate server exists
	if err := s.ValidateAppServerExists(app.ServerID); err != nil {
		return err
	}

	// Set default status if not provided
	if app.Status == "" {
		app.Status = "inactive" // Will be updated to active if deploy succeeds? Or keep inactive?
	}

	// Save app to DB first
	if err := s.repo.Create(app); err != nil {
		return err
	}

	// Fetch server details (Username, IP, Password) needed for SSH
	// CreateApp might not load relation, so fetch server explicitly or rely on repo preloading if available.
	// But `app.Server` might be empty.
	server, err := s.serverRepo.FindByID(app.ServerID)
	if err != nil {
		return err
	}
	app.Server = *server

	// Deploy to ArgoCD
	if err := s.automationService.DeployArgoCDApp(app); err != nil {
		// Log error but don't fail the creation? Or fail?
		// User probably wants to know.
		return fmt.Errorf("app created but failed to deploy to ArgoCD: %w", err)
	}

	// Update status to active
	app.Status = "active"
	if err := s.repo.Update(app); err != nil {
		log.Printf("Failed to update app status to active: %v", err)
	}

	return nil
}

func (s *appService) GetAppByID(id uint) (*model.App, error) {
	return s.repo.FindByID(id)
}

func (s *appService) GetAppsByServerID(serverID uint) ([]model.App, error) {
	return s.repo.FindByServerID(serverID)
}

func (s *appService) GetAllApps() ([]model.App, error) {
	return s.repo.FindAll()
}

func (s *appService) UpdateApp(app *model.App) error {
	return s.repo.Update(app)
}

func (s *appService) DeleteApp(id uint) error {
	return s.repo.Delete(id)
}

func (s *appService) ValidateAppServerExists(serverID uint) error {
	server, err := s.serverRepo.FindByID(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		return errors.New("server not found")
	}
	return nil
}
