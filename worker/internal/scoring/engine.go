package scoring

import (
	"context"
	"math"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/rs/zerolog/log"
)

// Engine orchestrates all fraud scoring heuristics and writes the final score.
type Engine struct {
	driver neo4j.DriverWithContext
}

// NewEngine creates a new scoring engine.
func NewEngine(driver neo4j.DriverWithContext) *Engine {
	return &Engine{driver: driver}
}

// ScoreWallet runs all heuristics against a wallet and updates its risk_score.
func (e *Engine) ScoreWallet(ctx context.Context, address string) error {
	var totalScore float64
	factors := make(map[string]float64)

	// Run each heuristic
	if score, err := e.detectCycles(ctx, address); err != nil {
		log.Warn().Err(err).Str("address", address).Msg("Cycle detection failed")
	} else if score > 0 {
		factors["cycles"] = score
		totalScore += score
	}

	if score, err := e.detectFanOut(ctx, address); err != nil {
		log.Warn().Err(err).Str("address", address).Msg("Fan-out detection failed")
	} else if score > 0 {
		factors["fan_out"] = score
		totalScore += score
	}

	if score, err := e.detectFanIn(ctx, address); err != nil {
		log.Warn().Err(err).Str("address", address).Msg("Fan-in detection failed")
	} else if score > 0 {
		factors["fan_in"] = score
		totalScore += score
	}

	if score, err := e.detectVelocityAnomaly(ctx, address); err != nil {
		log.Warn().Err(err).Str("address", address).Msg("Velocity anomaly detection failed")
	} else if score > 0 {
		factors["velocity"] = score
		totalScore += score
	}

	if score, err := e.detectProximityToBlacklisted(ctx, address); err != nil {
		log.Warn().Err(err).Str("address", address).Msg("Proximity detection failed")
	} else if score > 0 {
		factors["proximity"] = score
		totalScore += score
	}

	if score, err := e.detectHighRiskProgramUsage(ctx, address); err != nil {
		log.Warn().Err(err).Str("address", address).Msg("Program risk detection failed")
	} else if score > 0 {
		factors["program_risk"] = score
		totalScore += score
	}

	// Clamp to 0-100, normalize to 0.0-1.0
	totalScore = math.Min(totalScore, 100)
	normalizedScore := totalScore / 100.0

	// Write score back to wallet
	return e.writeScore(ctx, address, normalizedScore, factors)
}

func (e *Engine) writeScore(ctx context.Context, address string, score float64, factors map[string]float64) error {
	session := e.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// Convert factors to a list of strings for storage
		var factorList []string
		for name, score := range factors {
			factorList = append(factorList, name)
			_ = score // Score details could be stored in a more structured way later
		}

		query := `
			MATCH (w:Wallet {address: $address})
			SET w.risk_score = $score,
			    w.risk_factors = $factors,
			    w.scored_at = datetime()
		`
		_, err := tx.Run(ctx, query, map[string]interface{}{
			"address": address,
			"score":   score,
			"factors": factorList,
		})
		return nil, err
	})

	if err == nil {
		log.Info().Str("address", address).Float64("score", score).Interface("factors", factors).Msg("Scored wallet")
	}
	return err
}
