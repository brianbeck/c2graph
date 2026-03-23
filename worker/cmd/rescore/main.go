package main

import (
	"context"
	"os"
	"time"

	"github.com/brianbeck/c2graph-worker/internal/config"
	"github.com/brianbeck/c2graph-worker/internal/scoring"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// rescore is a one-shot command that scores all existing wallets in Neo4j.
// Usage: go run ./cmd/rescore
// Or inside Docker: docker compose exec worker /usr/local/bin/c2graph-rescore
func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	log.Info().Msg("Starting wallet rescore")

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	driver, err := neo4j.NewDriverWithContext(cfg.Neo4jURI, neo4j.BasicAuth(cfg.Neo4jUser, cfg.Neo4jPassword, ""))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Neo4j driver")
	}
	defer driver.Close(context.Background())

	ctx := context.Background()
	if err := driver.VerifyConnectivity(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Neo4j")
	}

	// Get all wallet addresses
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		rec, err := tx.Run(ctx, `MATCH (w:Wallet) RETURN w.address AS address`, nil)
		if err != nil {
			return nil, err
		}
		var addrs []string
		for rec.Next(ctx) {
			if val, ok := rec.Record().Get("address"); ok {
				if addr, ok := val.(string); ok {
					addrs = append(addrs, addr)
				}
			}
		}
		return addrs, nil
	})
	session.Close(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to fetch wallets")
	}

	addrs := result.([]string)
	log.Info().Int("wallets", len(addrs)).Msg("Found wallets to score")

	scorer := scoring.NewEngine(driver)
	scored, failed := 0, 0

	for i, addr := range addrs {
		if err := scorer.ScoreWallet(ctx, addr); err != nil {
			log.Warn().Err(err).Str("address", addr).Msg("Failed to score wallet")
			failed++
		} else {
			scored++
		}

		if (i+1)%50 == 0 {
			log.Info().Int("progress", i+1).Int("total", len(addrs)).Int("scored", scored).Int("failed", failed).Msg("Rescore progress")
		}
	}

	log.Info().Int("scored", scored).Int("failed", failed).Int("total", len(addrs)).Msg("Rescore complete")
}
