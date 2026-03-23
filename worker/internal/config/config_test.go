package config

import (
	"os"
	"testing"
)

func clearSolanaEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{"SOLANA_RPC_URL", "SOLANA_RPC_RATE_LIMIT", "SOLANA_RPC_URLS", "SOLANA_RPC_RATE_LIMITS"} {
		os.Unsetenv(key)
	}
}

func TestLoad_Defaults(t *testing.T) {
	for _, key := range []string{"NEO4J_URI", "NEO4J_USER", "NEO4J_PASSWORD", "RABBITMQ_URI",
		"SOLANA_RPC_URL", "SOLANA_RPC_RATE_LIMIT", "SOLANA_RPC_URLS", "SOLANA_RPC_RATE_LIMITS",
		"WORKER_CONCURRENCY", "BATCH_SIZE", "MAX_WALLETS_PER_SCAN", "SCAN_FRESHNESS_HOURS"} {
		os.Unsetenv(key)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.Neo4jURI != "bolt://localhost:7687" {
		t.Errorf("Neo4jURI = %q, want %q", cfg.Neo4jURI, "bolt://localhost:7687")
	}
	if len(cfg.SolanaRPCURLs) != 1 || cfg.SolanaRPCURLs[0] != "https://api.mainnet-beta.solana.com" {
		t.Errorf("SolanaRPCURLs = %v", cfg.SolanaRPCURLs)
	}
	if len(cfg.SolanaRPCRateLimits) != 1 || cfg.SolanaRPCRateLimits[0] != 2 {
		t.Errorf("SolanaRPCRateLimits = %v", cfg.SolanaRPCRateLimits)
	}
	if cfg.Concurrency != 4 {
		t.Errorf("Concurrency = %d, want 4", cfg.Concurrency)
	}
	if cfg.BatchSize != 50 {
		t.Errorf("BatchSize = %d, want 50", cfg.BatchSize)
	}
	if cfg.MaxWalletsPerScan != 10000 {
		t.Errorf("MaxWalletsPerScan = %d, want 10000", cfg.MaxWalletsPerScan)
	}
	if cfg.ScanFreshnessHrs != 24 {
		t.Errorf("ScanFreshnessHrs = %d, want 24", cfg.ScanFreshnessHrs)
	}
}

func TestLoad_LegacySingleEndpoint(t *testing.T) {
	clearSolanaEnv(t)
	t.Setenv("SOLANA_RPC_URL", "https://my-rpc.example.com")
	t.Setenv("SOLANA_RPC_RATE_LIMIT", "10")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if len(cfg.SolanaRPCURLs) != 1 {
		t.Fatalf("expected 1 URL, got %d", len(cfg.SolanaRPCURLs))
	}
	if cfg.SolanaRPCURLs[0] != "https://my-rpc.example.com" {
		t.Errorf("URL = %q", cfg.SolanaRPCURLs[0])
	}
	if cfg.SolanaRPCRateLimits[0] != 10 {
		t.Errorf("rate = %d, want 10", cfg.SolanaRPCRateLimits[0])
	}
}

func TestLoad_MultiEndpoint(t *testing.T) {
	clearSolanaEnv(t)
	t.Setenv("SOLANA_RPC_URLS", "https://helius.example.com, https://api.mainnet-beta.solana.com")
	t.Setenv("SOLANA_RPC_RATE_LIMITS", "10,2")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if len(cfg.SolanaRPCURLs) != 2 {
		t.Fatalf("expected 2 URLs, got %d", len(cfg.SolanaRPCURLs))
	}
	if cfg.SolanaRPCURLs[0] != "https://helius.example.com" {
		t.Errorf("URL[0] = %q", cfg.SolanaRPCURLs[0])
	}
	if cfg.SolanaRPCURLs[1] != "https://api.mainnet-beta.solana.com" {
		t.Errorf("URL[1] = %q", cfg.SolanaRPCURLs[1])
	}
	if cfg.SolanaRPCRateLimits[0] != 10 {
		t.Errorf("rate[0] = %d", cfg.SolanaRPCRateLimits[0])
	}
	if cfg.SolanaRPCRateLimits[1] != 2 {
		t.Errorf("rate[1] = %d", cfg.SolanaRPCRateLimits[1])
	}
}

func TestLoad_MultiEndpoint_RatePadding(t *testing.T) {
	clearSolanaEnv(t)
	t.Setenv("SOLANA_RPC_URLS", "https://a.com,https://b.com,https://c.com")
	t.Setenv("SOLANA_RPC_RATE_LIMITS", "10")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if len(cfg.SolanaRPCRateLimits) != 3 {
		t.Fatalf("expected 3 rates, got %d", len(cfg.SolanaRPCRateLimits))
	}
	if cfg.SolanaRPCRateLimits[0] != 10 {
		t.Errorf("rate[0] = %d, want 10", cfg.SolanaRPCRateLimits[0])
	}
	if cfg.SolanaRPCRateLimits[1] != 2 {
		t.Errorf("rate[1] = %d, want 2 (padded default)", cfg.SolanaRPCRateLimits[1])
	}
}

func TestLoad_MultiEndpointOverridesLegacy(t *testing.T) {
	clearSolanaEnv(t)
	t.Setenv("SOLANA_RPC_URL", "https://legacy.example.com")
	t.Setenv("SOLANA_RPC_URLS", "https://primary.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	// Multi-endpoint takes precedence
	if len(cfg.SolanaRPCURLs) != 1 {
		t.Fatalf("expected 1 URL, got %d", len(cfg.SolanaRPCURLs))
	}
	if cfg.SolanaRPCURLs[0] != "https://primary.example.com" {
		t.Errorf("URL = %q, want primary not legacy", cfg.SolanaRPCURLs[0])
	}
}

func TestValidate_Errors(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name:    "empty Neo4jURI",
			cfg:     Config{Neo4jURI: "", RabbitMQURI: "amqp://", SolanaRPCURLs: []string{"https://"}, Concurrency: 1, MaxWalletsPerScan: 1},
			wantErr: "NEO4J_URI is required",
		},
		{
			name:    "empty RabbitMQURI",
			cfg:     Config{Neo4jURI: "bolt://", RabbitMQURI: "", SolanaRPCURLs: []string{"https://"}, Concurrency: 1, MaxWalletsPerScan: 1},
			wantErr: "RABBITMQ_URI is required",
		},
		{
			name:    "empty SolanaRPCURLs",
			cfg:     Config{Neo4jURI: "bolt://", RabbitMQURI: "amqp://", SolanaRPCURLs: nil, Concurrency: 1, MaxWalletsPerScan: 1},
			wantErr: "SOLANA_RPC_URL or SOLANA_RPC_URLS is required",
		},
		{
			name:    "zero concurrency",
			cfg:     Config{Neo4jURI: "bolt://", RabbitMQURI: "amqp://", SolanaRPCURLs: []string{"https://"}, Concurrency: 0, MaxWalletsPerScan: 1},
			wantErr: "WORKER_CONCURRENCY must be >= 1",
		},
		{
			name:    "zero MaxWalletsPerScan",
			cfg:     Config{Neo4jURI: "bolt://", RabbitMQURI: "amqp://", SolanaRPCURLs: []string{"https://"}, Concurrency: 1, MaxWalletsPerScan: 0},
			wantErr: "MAX_WALLETS_PER_SCAN must be >= 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.validate()
			if err == nil {
				t.Fatal("expected error")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestValidate_OK(t *testing.T) {
	cfg := Config{
		Neo4jURI:          "bolt://localhost:7687",
		RabbitMQURI:       "amqp://localhost:5672/",
		SolanaRPCURLs:     []string{"https://api.mainnet-beta.solana.com"},
		Concurrency:       4,
		MaxWalletsPerScan: 10000,
	}
	if err := cfg.validate(); err != nil {
		t.Errorf("validate() unexpected error: %v", err)
	}
}
