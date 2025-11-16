package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"log"

	"github.com/melbahja/goph"
)

type AutomationService struct {
	clusterRepo *repository.ClusterRepository
}

func NewAutomationService(clusterRepo *repository.ClusterRepository) *AutomationService {
	return &AutomationService{
		clusterRepo: clusterRepo,
	}
}

func (s *AutomationService) AutoInstallK3sToServer(cluster *model.Cluster) error {
	username := cluster.Username
	ip := cluster.IpAddress
	password := cluster.Password
	client, err := goph.New(username, ip, goph.Password(password))

	log.Printf("Server connected: %v", client)

	if err != nil {
		return err
	}

	defer client.Close()

	k3s, err := client.Run("k3s --version")

	if err != nil {
		return err
	}

	if k3s != nil {
		log.Printf("Server response: %v", k3s)
		return nil
	}

	installCmd := "curl -sfL https://get.k3s.io | sh -"
	_, err = client.Run(installCmd)

	if err != nil {
		return err
	}

	log.Printf("K3s installed successfully on server: %s", ip)

	client.Run("mkdir /var/lib/app")

	err = client.Upload("/lib/helm", "/var/lib/app")

	if err != nil {
		return err
	}

	log.Printf("uploaded config helm chart on server: %s", ip)

	return nil
}
