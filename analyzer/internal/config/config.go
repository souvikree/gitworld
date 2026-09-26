package config

import (
	"fmt"
	"os"
)

type Config struct {
	Neo4jURI      string
	Neo4jUser     string
	Neo4jPassword string
	RedisURL      string
	Port          string
	LogLevel      string
}

func Load() (*Config, error) {
	cfg := &Config{
		Neo4jURI:      os.Getenv("NEO4J_URI"),
		Neo4jUser:     os.Getenv("NEO4J_USER"),
		Neo4jPassword: os.Getenv("NEO4J_PASSWORD"),
		RedisURL:      os.Getenv("REDIS_URL"),
		Port:          os.Getenv("API_PORT"),
		LogLevel:      os.Getenv("LOG_LEVEL"),
	}

	if cfg.Neo4jURI == "" {
		return nil, fmt.Errorf("config: NEO4J_URI is required")
	}
	if cfg.Port == "" {
		cfg.Port = "8080" // sane default
	}

	return cfg, nil
}