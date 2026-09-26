package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Neo4jURI      string
	Neo4jUser     string
	Neo4jPassword string
	Port          string
	LogLevel      string
	APIKey        string // for the auth middleware, added next step
}

func Load() (*Config, error) {
	_ = godotenv.Load("../infra/.env")

	cfg := &Config{
		Neo4jURI:      os.Getenv("NEO4J_URI"),
		Neo4jUser:     os.Getenv("NEO4J_USER"),
		Neo4jPassword: os.Getenv("NEO4J_PASSWORD"),
		Port:          os.Getenv("API_PORT"),
		LogLevel:      os.Getenv("LOG_LEVEL"),
		APIKey:        os.Getenv("API_KEY"),
	}

	if cfg.Neo4jURI == "" {
		return nil, fmt.Errorf("config: NEO4J_URI is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("config: API_KEY is required")
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg, nil
}