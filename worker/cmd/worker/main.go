package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brianbeck/sentinel-worker/internal/config"
	"github.com/brianbeck/sentinel-worker/internal/consumer"
	neo4jstore "github.com/brianbeck/sentinel-worker/internal/neo4j"
	"github.com/brianbeck/sentinel-worker/internal/solana"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Pretty console logging for development
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	log.Info().Msg("Starting Sentinel Worker")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Connect to Neo4j
	driver, err := neo4j.NewDriverWithContext(cfg.Neo4jURI, neo4j.BasicAuth(cfg.Neo4jUser, cfg.Neo4jPassword, ""))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Neo4j driver")
	}
	defer driver.Close(context.Background())

	// Verify Neo4j connectivity
	ctx := context.Background()
	if err := driver.VerifyConnectivity(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Neo4j")
	}
	log.Info().Str("uri", cfg.Neo4jURI).Msg("Connected to Neo4j")

	// Create services
	solClient := solana.NewClient(cfg.SolanaRPCURL, cfg.SolanaRPCRateLimit)
	writer := neo4jstore.NewWriter(driver)
	dedup := neo4jstore.NewDedupChecker(driver, cfg.ScanFreshnessHrs)

	// Create consumer
	cons := consumer.NewConsumer(cfg, solClient, writer, dedup)
	defer cons.Close()

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Info().Str("signal", sig.String()).Msg("Shutdown signal received")
		cancel()
	}()

	// Start consuming
	if err := cons.Start(ctx); err != nil {
		if ctx.Err() != nil {
			log.Info().Msg("Worker shutting down gracefully")
		} else {
			log.Fatal().Err(err).Msg("Worker failed")
		}
	}
}
