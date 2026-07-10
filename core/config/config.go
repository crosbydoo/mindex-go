package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port          string
	PostgresURL   string
	AdminPassword string
	CORSOrigin    string
}

func Load() (*Config, error) {
	postgresURL := os.Getenv("POSTGRES_URL")
	if postgresURL == "" {
		return nil, fmt.Errorf("POSTGRES_URL is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		Port:          port,
		PostgresURL:   postgresURL,
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		CORSOrigin:    os.Getenv("CORS_ORIGIN"),
	}, nil
}
