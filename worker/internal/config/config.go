package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all worker configuration loaded from environment variables.
type Config struct {
	// Neo4j
	Neo4jURI      string
	Neo4jUser     string
	Neo4jPassword string

	// RabbitMQ
	RabbitMQURI string

	// Solana RPC — multiple endpoints with per-endpoint rate limits.
	// SOLANA_RPC_URLS is a comma-separated list (primary first, fallbacks after).
	// SOLANA_RPC_RATE_LIMITS is a matching comma-separated list of req/sec.
	// Legacy SOLANA_RPC_URL / SOLANA_RPC_RATE_LIMIT are still supported as the first entry.
	SolanaRPCURLs       []string
	SolanaRPCRateLimits []int

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
		Neo4jPassword:     getEnv("NEO4J_PASSWORD", "c2graph-dev-password"),
		RabbitMQURI:       getEnv("RABBITMQ_URI", "amqp://guest:guest@localhost:5672/"),
		Concurrency:       getEnvInt("WORKER_CONCURRENCY", 4),
		BatchSize:         getEnvInt("BATCH_SIZE", 50),
		MaxWalletsPerScan: getEnvInt("MAX_WALLETS_PER_SCAN", 10000),
		ScanFreshnessHrs:  getEnvInt("SCAN_FRESHNESS_HOURS", 24),
	}

	// Parse RPC endpoints. Prefer the new multi-endpoint vars, fall back to legacy single.
	c.SolanaRPCURLs, c.SolanaRPCRateLimits = parseRPCEndpoints()

	if err := c.validate(); err != nil {
		return nil, err
	}
	return c, nil
}

// parseRPCEndpoints reads SOLANA_RPC_URLS + SOLANA_RPC_RATE_LIMITS (comma-separated)
// and falls back to SOLANA_RPC_URL + SOLANA_RPC_RATE_LIMIT (single endpoint).
func parseRPCEndpoints() ([]string, []int) {
	var urls []string
	var rates []int

	// New multi-endpoint config
	if raw := os.Getenv("SOLANA_RPC_URLS"); raw != "" {
		for _, u := range strings.Split(raw, ",") {
			u = strings.TrimSpace(u)
			if u != "" {
				urls = append(urls, u)
			}
		}
	}

	if raw := os.Getenv("SOLANA_RPC_RATE_LIMITS"); raw != "" {
		for _, r := range strings.Split(raw, ",") {
			r = strings.TrimSpace(r)
			v, err := strconv.Atoi(r)
			if err != nil || v < 1 {
				v = 2
			}
			rates = append(rates, v)
		}
	}

	// Fall back to legacy single-endpoint vars
	if len(urls) == 0 {
		u := getEnv("SOLANA_RPC_URL", "https://api.mainnet-beta.solana.com")
		urls = append(urls, u)
		rates = append(rates, getEnvInt("SOLANA_RPC_RATE_LIMIT", 2))
	}

	// Pad rates to match urls length
	for len(rates) < len(urls) {
		rates = append(rates, 2)
	}

	return urls, rates
}

func (c *Config) validate() error {
	if c.Neo4jURI == "" {
		return fmt.Errorf("NEO4J_URI is required")
	}
	if c.RabbitMQURI == "" {
		return fmt.Errorf("RABBITMQ_URI is required")
	}
	if len(c.SolanaRPCURLs) == 0 {
		return fmt.Errorf("SOLANA_RPC_URL or SOLANA_RPC_URLS is required")
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
