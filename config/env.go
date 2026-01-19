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
	RedisUrl           string
	AppEnv             string
}

func LoadEnv() *EnvConfig {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, continuing with environment variables")
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
		RedisUrl:           os.Getenv("REDIS_URL"),
		AppEnv:             os.Getenv("APP_ENV"),
	}
}
