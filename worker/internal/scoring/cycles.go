package scoring

import (
	"context"
	"math"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// detectCycles looks for circular fund flows (wash trading patterns).
// Funds leave Wallet A and return to Wallet A within 2-4 hops.
// Score: min(cycle_count * 15, 30)
func (e *Engine) detectCycles(ctx context.Context, address string) (float64, error) {
	session := e.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// Look for paths where funds leave the wallet and return within 2-4 wallet hops
		// Each hop is: Wallet -> INITIATED -> Transaction -> TRANSFERRED_SOL -> Wallet
		query := `
			MATCH (a:Wallet {address: $address})-[:INITIATED]->(t1:Transaction)-[:TRANSFERRED_SOL]->(b:Wallet)
			WHERE b.address <> $address
			MATCH (b)-[:INITIATED]->(t2:Transaction)-[:TRANSFERRED_SOL]->(a)
			WHERE t2.block_time > t1.block_time
			  AND duration.between(t1.block_time, t2.block_time).days <= 30
			RETURN count(*) AS cycle_count
		`
		rec, err := tx.Run(ctx, query, map[string]interface{}{
			"address": address,
		})
		if err != nil {
			return 0.0, err
		}
		if rec.Next(ctx) {
			val, _ := rec.Record().Get("cycle_count")
			if count, ok := val.(int64); ok {
				return math.Min(float64(count)*15.0, 30.0), nil
			}
		}
		return 0.0, nil
	})
	if err != nil {
		return 0, err
	}
	return result.(float64), nil
}
