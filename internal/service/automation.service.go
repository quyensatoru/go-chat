package service

import (
	"backend/internal/repository"
)

type AutomationService struct {
	ClusterRepository *repository.ClusterRepository
}

func NewAutomationService(clusterRepo *repository.ClusterRepository) *AutomationService {
	return &AutomationService{
		ClusterRepository: clusterRepo,
	}
}
