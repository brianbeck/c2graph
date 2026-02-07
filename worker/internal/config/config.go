package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all worker configuration loaded from environment variables.
type Config struct {
	// Neo4j
	Neo4jURI      string
	Neo4jUser     string
	Neo4jPassword string

	// RabbitMQ
	RabbitMQURI string

	// Solana
	SolanaRPCURL      string
	SolanaRPCRateLimit int // requests per second

	// Worker behavior
	Concurrency       int
	BatchSize         int
	MaxWalletsPerScan int
	ScanFreshnessHrs  int
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	c := &Config{
		Neo4jURI:          getEnv("NEO4J_URI", "bolt://localhost:7687"),
		Neo4jUser:         getEnv("NEO4J_USER", "neo4j"),
		Neo4jPassword:     getEnv("NEO4J_PASSWORD", "sentinel-dev-password"),
		RabbitMQURI:       getEnv("RABBITMQ_URI", "amqp://guest:guest@localhost:5672/"),
		SolanaRPCURL:      getEnv("SOLANA_RPC_URL", "https://api.mainnet-beta.solana.com"),
		SolanaRPCRateLimit: getEnvInt("SOLANA_RPC_RATE_LIMIT", 2),
		Concurrency:       getEnvInt("WORKER_CONCURRENCY", 4),
		BatchSize:         getEnvInt("BATCH_SIZE", 50),
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
	if c.SolanaRPCURL == "" {
		return fmt.Errorf("SOLANA_RPC_URL is required")
	}
	if c.Concurrency < 1 {
		return fmt.Errorf("WORKER_CONCURRENCY must be >= 1")
	}
	if c.MaxWalletsPerScan < 1 {
		return fmt.Errorf("MAX_WALLETS_PER_SCAN must be >= 1")
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
