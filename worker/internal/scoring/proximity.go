package scoring

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// detectProximityToBlacklisted checks distance to known blacklisted wallets.
// 1 hop → +50, 2 hops → +20, 3 hops → +5
func (e *Engine) detectProximityToBlacklisted(ctx context.Context, address string) (float64, error) {
	session := e.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// Find shortest path to any blacklisted wallet
		query := `
			MATCH (a:Wallet {address: $address}), (b:Wallet)
			WHERE 'Blacklisted' IN COALESCE(b.tags, [])
			  AND a <> b
			MATCH path = shortestPath((a)-[*..6]-(b))
			WITH length(path) AS distance
			ORDER BY distance
			LIMIT 1
			RETURN distance
		`
		rec, err := tx.Run(ctx, query, map[string]interface{}{
			"address": address,
		})
		if err != nil {
			return 0.0, err
		}
		if rec.Next(ctx) {
			val, _ := rec.Record().Get("distance")
			if dist, ok := val.(int64); ok {
				// Convert graph distance to wallet hops
				// In our model: Wallet -> Transaction -> Wallet = distance 2 = 1 wallet hop
				walletHops := dist / 2
				switch {
				case walletHops <= 1:
					return 50.0, nil
				case walletHops <= 2:
					return 20.0, nil
				case walletHops <= 3:
					return 5.0, nil
				}
			}
		}
		return 0.0, nil
	})
	if err != nil {
		return 0, err
	}
	return result.(float64), nil
}

// detectHighRiskProgramUsage checks if the wallet interacts with known high-risk programs.
// Score: +15 per high-risk program
func (e *Engine) detectHighRiskProgramUsage(ctx context.Context, address string) (float64, error) {
	session := e.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (w:Wallet {address: $address})-[:INITIATED]->(:Transaction)-[:CONTAINS_INSTRUCTION]->(:Instruction)-[:EXECUTED_BY]->(p:Program)
			WHERE p.risk_level = 'high'
			RETURN count(DISTINCT p) AS high_risk_count
		`
		rec, err := tx.Run(ctx, query, map[string]interface{}{
			"address": address,
		})
		if err != nil {
			return 0.0, err
		}
		if rec.Next(ctx) {
			val, _ := rec.Record().Get("high_risk_count")
			if count, ok := val.(int64); ok && count > 0 {
				return float64(count) * 15.0, nil
			}
		}
		return 0.0, nil
	})
	if err != nil {
		return 0, err
	}
	return result.(float64), nil
}
