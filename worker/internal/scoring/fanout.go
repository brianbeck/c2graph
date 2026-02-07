package scoring

import (
	"context"
	"math"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// detectFanOut checks if a wallet distributes funds to 10+ unique wallets within 1-hour windows.
// This is a structuring pattern used to split large amounts.
// Score: min(fanout_windows * 12, 25)
func (e *Engine) detectFanOut(ctx context.Context, address string) (float64, error) {
	session := e.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (w:Wallet {address: $address})-[:INITIATED]->(t:Transaction)-[:TRANSFERRED_SOL]->(r:Wallet)
			WHERE r.address <> $address AND t.block_time IS NOT NULL
			WITH w, t.block_time AS ts, r
			WITH w, datetime.truncate('hour', ts) AS hour_bucket, collect(DISTINCT r.address) AS recipients
			WHERE size(recipients) >= 10
			RETURN count(*) AS fanout_windows
		`
		rec, err := tx.Run(ctx, query, map[string]interface{}{
			"address": address,
		})
		if err != nil {
			return 0.0, err
		}
		if rec.Next(ctx) {
			val, _ := rec.Record().Get("fanout_windows")
			if count, ok := val.(int64); ok {
				return math.Min(float64(count)*12.0, 25.0), nil
			}
		}
		return 0.0, nil
	})
	if err != nil {
		return 0, err
	}
	return result.(float64), nil
}

// detectFanIn checks if a wallet receives funds from 10+ unique wallets within 1-hour windows.
// This is a consolidation pattern.
// Score: min(fanin_windows * 12, 25)
func (e *Engine) detectFanIn(ctx context.Context, address string) (float64, error) {
	session := e.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (sender:Wallet)-[:INITIATED]->(t:Transaction)-[:TRANSFERRED_SOL]->(w:Wallet {address: $address})
			WHERE sender.address <> $address AND t.block_time IS NOT NULL
			WITH w, t.block_time AS ts, sender
			WITH w, datetime.truncate('hour', ts) AS hour_bucket, collect(DISTINCT sender.address) AS senders
			WHERE size(senders) >= 10
			RETURN count(*) AS fanin_windows
		`
		rec, err := tx.Run(ctx, query, map[string]interface{}{
			"address": address,
		})
		if err != nil {
			return 0.0, err
		}
		if rec.Next(ctx) {
			val, _ := rec.Record().Get("fanin_windows")
			if count, ok := val.(int64); ok {
				return math.Min(float64(count)*12.0, 25.0), nil
			}
		}
		return 0.0, nil
	})
	if err != nil {
		return 0, err
	}
	return result.(float64), nil
}
