package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all API gateway configuration.
type Config struct {
	// Server
	Port       string
	CORSOrigin string

	// Neo4j
	Neo4jURI      string
	Neo4jUser     string
	Neo4jPassword string

	// RabbitMQ
	RabbitMQURI string

	// Supabase Auth
	SupabaseURL       string
	SupabaseJWTSecret string

	// Behavior
	MaxWalletsPerScan int
	ScanFreshnessHrs  int
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	c := &Config{
		Port:              getEnv("API_PORT", "8080"),
		CORSOrigin:        getEnv("API_CORS_ORIGIN", "http://localhost:5173"),
		Neo4jURI:          getEnv("NEO4J_URI", "bolt://localhost:7687"),
		Neo4jUser:         getEnv("NEO4J_USER", "neo4j"),
		Neo4jPassword:     getEnv("NEO4J_PASSWORD", "sentinel-dev-password"),
		RabbitMQURI:       getEnv("RABBITMQ_URI", "amqp://guest:guest@localhost:5672/"),
		SupabaseURL:       getEnv("SUPABASE_URL", ""),
		SupabaseJWTSecret: getEnv("SUPABASE_JWT_SECRET", ""),
		MaxWalletsPerScan: getEnvInt("MAX_WALLETS_PER_SCAN", 10000),
		ScanFreshnessHrs:  getEnvInt("SCAN_FRESHNESS_HOURS", 24),
	}

	if err := c.validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Config) validate() error {
	if c.Neo4jURI == "" {
		return fmt.Errorf("NEO4J_URI is required")
	}
	if c.RabbitMQURI == "" {
		return fmt.Errorf("RABBITMQ_URI is required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}
