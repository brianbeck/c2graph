package scoring

import (
	"context"
	"math"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// detectVelocityAnomaly checks if a wallet's transaction frequency is abnormally high
// relative to its age. Baseline: 5 tx/day. >5x baseline = +20.
func (e *Engine) detectVelocityAnomaly(ctx context.Context, address string) (float64, error) {
	session := e.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (w:Wallet {address: $address})
			WHERE w.first_seen IS NOT NULL AND w.tx_count IS NOT NULL
			WITH w,
			     w.tx_count AS tx_count,
			     duration.between(w.first_seen, datetime()).days AS age_days
			WHERE age_days > 0
			WITH tx_count, age_days, toFloat(tx_count) / toFloat(age_days) AS tx_per_day
			RETURN tx_per_day, tx_count, age_days
		`
		rec, err := tx.Run(ctx, query, map[string]interface{}{
			"address": address,
		})
		if err != nil {
			return 0.0, err
		}
		if rec.Next(ctx) {
			val, _ := rec.Record().Get("tx_per_day")
			var txPerDay float64
			switch v := val.(type) {
			case float64:
				txPerDay = v
			case int64:
				txPerDay = float64(v)
			}

			// Baseline: 5 tx/day is considered normal
			const baseline = 5.0
			if txPerDay > baseline*5 {
				// Scale score: 5x = 10 points, 10x = 20 points (capped)
				multiplier := txPerDay / baseline
				score := math.Min(multiplier*2.0, 20.0)
				return score, nil
			}
		}
		return 0.0, nil
	})
	if err != nil {
		return 0, err
	}
	return result.(float64), nil
}
