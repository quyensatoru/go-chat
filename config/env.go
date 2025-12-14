package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type EnvConfig struct {
	Port               string
	DBHost             string
	DBUser             string
	DBPassword         string
	DBName             string
	DBPort             string
	GitOpsRepo         string
	GitlabApiUrl       string
	GitlabPrivateToken string
	GitlabProjectID    string
	GitOpsToken        string
}

func LoadEnv() *EnvConfig {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return &EnvConfig{
		Port:               os.Getenv("PORT"),
		DBHost:             os.Getenv("DB_HOST"),
		DBUser:             os.Getenv("DB_USER"),
		DBPassword:         os.Getenv("DB_PASSWORD"),
		DBName:             os.Getenv("DB_NAME"),
		DBPort:             os.Getenv("DB_PORT"),
		GitOpsRepo:         os.Getenv("GITOPS_REPO"),
		GitlabApiUrl:       os.Getenv("GITLAB_API_URL"),
		GitlabPrivateToken: os.Getenv("GITLAB_PRIVATE_TOKEN"),
		GitlabProjectID:    os.Getenv("GITLAB_PROJECT_ID"),
		GitOpsToken:        os.Getenv("GITOPS_TOKEN"),
	}
}
