package scoring

import (
	"context"
	"math"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// botSignals collects individual bot detection signal scores.
type botSignals struct {
	timingRegularity   float64 // Low stddev of inter-tx intervals
	velocityScore      float64 // Very high tx/day
	programConcentrate float64 // Uses very few programs repeatedly
	hourSpread         float64 // Active across all 24 hours (no sleep)
}

// BotResult holds the bot detection outcome with individual signal scores.
type BotResult struct {
	IsBot              bool
	Likelihood         float64
	RiskContribution   float64
	TimingRegularity   float64
	Velocity           float64
	ProgramConcentrate float64
	HourSpread         float64
}

// detectBotBehavior runs multiple heuristics to determine if a wallet is likely a bot.
// Returns a BotResult with the overall likelihood, individual signal scores, and risk contribution.
func (e *Engine) detectBotBehavior(ctx context.Context, address string) (*BotResult, error) {
	signals := botSignals{}

	if s, err := e.botTimingRegularity(ctx, address); err != nil {
		return nil, err
	} else {
		signals.timingRegularity = s
	}

	if s, err := e.botVelocity(ctx, address); err != nil {
		return nil, err
	} else {
		signals.velocityScore = s
	}

	if s, err := e.botProgramConcentration(ctx, address); err != nil {
		return nil, err
	} else {
		signals.programConcentrate = s
	}

	if s, err := e.botHourSpread(ctx, address); err != nil {
		return nil, err
	} else {
		signals.hourSpread = s
	}

	// Combine signals: each is 0.0-1.0, weighted sum
	likelihood := signals.timingRegularity*0.30 +
		signals.velocityScore*0.30 +
		signals.programConcentrate*0.20 +
		signals.hourSpread*0.20

	likelihood = math.Min(likelihood, 1.0)

	return &BotResult{
		IsBot:              likelihood >= 0.5,
		Likelihood:         likelihood,
		RiskContribution:   likelihood * 25.0,
		TimingRegularity:   signals.timingRegularity,
		Velocity:           signals.velocityScore,
		ProgramConcentrate: signals.programConcentrate,
		HourSpread:         signals.hourSpread,
	}, nil
}

// botTimingRegularity measures how regular the spacing between transactions is.
// Bots tend to have very consistent intervals. Returns 0.0-1.0.
func (e *Engine) botTimingRegularity(ctx context.Context, address string) (float64, error) {
	session := e.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// Get the coefficient of variation (stddev/mean) of inter-transaction intervals.
		// A low CV means highly regular timing — bot-like.
		query := `
			MATCH (w:Wallet {address: $address})-[:INITIATED]->(t:Transaction)
			WHERE t.block_time IS NOT NULL
			WITH t ORDER BY t.block_time
			WITH collect(t.block_time) AS times
			WHERE size(times) >= 10
			WITH [i IN range(1, size(times)-1) |
			  duration.between(times[i-1], times[i]).seconds] AS intervals
			WITH intervals
			WHERE size(intervals) > 0
			WITH reduce(s = 0.0, x IN intervals | s + x) / size(intervals) AS mean,
			     intervals
			WHERE mean > 0
			WITH mean,
			     sqrt(reduce(s = 0.0, x IN intervals | s + (x - mean)*(x - mean)) / size(intervals)) AS stddev
			RETURN CASE WHEN mean > 0 THEN stddev / mean ELSE 1.0 END AS cv
		`
		rec, err := tx.Run(ctx, query, map[string]interface{}{
			"address": address,
		})
		if err != nil {
			return 0.0, err
		}
		if rec.Next(ctx) {
			val, _ := rec.Record().Get("cv")
			cv := toFloat(val)
			// CV < 0.1 = extremely regular (strong bot signal)
			// CV < 0.3 = moderately regular
			// CV > 0.5 = human-like variance
			if cv < 0.1 {
				return 1.0, nil
			} else if cv < 0.3 {
				return 0.7, nil
			} else if cv < 0.5 {
				return 0.3, nil
			}
		}
		return 0.0, nil
	})
	if err != nil {
		return 0, err
	}
	return result.(float64), nil
}

// botVelocity checks for extremely high transaction throughput.
// Returns 0.0-1.0 based on tx/day. >100/day is strongly bot-like.
func (e *Engine) botVelocity(ctx context.Context, address string) (float64, error) {
	session := e.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (w:Wallet {address: $address})
			WHERE w.first_seen IS NOT NULL AND w.tx_count IS NOT NULL
			WITH w.tx_count AS tx_count,
			     duration.between(w.first_seen, datetime()).days AS age_days
			WHERE age_days > 0
			RETURN toFloat(tx_count) / toFloat(age_days) AS tx_per_day
		`
		rec, err := tx.Run(ctx, query, map[string]interface{}{
			"address": address,
		})
		if err != nil {
			return 0.0, err
		}
		if rec.Next(ctx) {
			val, _ := rec.Record().Get("tx_per_day")
			txPerDay := toFloat(val)
			// >100 tx/day: very likely bot
			// >50 tx/day: probably bot
			// >25 tx/day: possibly bot
			if txPerDay > 100 {
				return 1.0, nil
			} else if txPerDay > 50 {
				return 0.7, nil
			} else if txPerDay > 25 {
				return 0.4, nil
			}
		}
		return 0.0, nil
	})
	if err != nil {
		return 0, err
	}
	return result.(float64), nil
}

// botProgramConcentration checks if the wallet only uses 1-2 programs across many transactions.
// Humans use a variety of programs; bots repeatedly call the same one.
// Returns 0.0-1.0.
func (e *Engine) botProgramConcentration(ctx context.Context, address string) (float64, error) {
	session := e.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (w:Wallet {address: $address})-[:INITIATED]->(t:Transaction)-[:CONTAINS_INSTRUCTION]->(i:Instruction)-[:EXECUTED_BY]->(p:Program)
			WHERE NOT p.program_id IN ['11111111111111111111111111111111', 'ComputeBudget111111111111111111111111111111']
			WITH w, count(DISTINCT t) AS tx_count, count(DISTINCT p) AS program_count
			WHERE tx_count >= 10
			RETURN tx_count, program_count,
			       toFloat(program_count) / toFloat(tx_count) AS diversity_ratio
		`
		rec, err := tx.Run(ctx, query, map[string]interface{}{
			"address": address,
		})
		if err != nil {
			return 0.0, err
		}
		if rec.Next(ctx) {
			programCount, _ := rec.Record().Get("program_count")
			txCount, _ := rec.Record().Get("tx_count")
			pc := toInt(programCount)
			tc := toInt(txCount)

			if tc >= 20 && pc <= 2 {
				return 1.0, nil // 20+ txns using only 1-2 programs
			} else if tc >= 10 && pc <= 2 {
				return 0.7, nil
			} else if tc >= 20 && pc <= 3 {
				return 0.4, nil
			}
		}
		return 0.0, nil
	})
	if err != nil {
		return 0, err
	}
	return result.(float64), nil
}

// botHourSpread checks if transactions are spread across all hours of the day.
// Humans typically don't transact 24/7. Bots do.
// Returns 0.0-1.0.
func (e *Engine) botHourSpread(ctx context.Context, address string) (float64, error) {
	session := e.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (w:Wallet {address: $address})-[:INITIATED]->(t:Transaction)
			WHERE t.block_time IS NOT NULL
			WITH DISTINCT t.block_time.hour AS hour
			RETURN count(hour) AS active_hours
		`
		rec, err := tx.Run(ctx, query, map[string]interface{}{
			"address": address,
		})
		if err != nil {
			return 0.0, err
		}
		if rec.Next(ctx) {
			val, _ := rec.Record().Get("active_hours")
			hours := toInt(val)
			// Active 22-24 hours: very likely bot
			// Active 18-21 hours: possibly bot
			if hours >= 22 {
				return 1.0, nil
			} else if hours >= 18 {
				return 0.6, nil
			} else if hours >= 15 {
				return 0.2, nil
			}
		}
		return 0.0, nil
	})
	if err != nil {
		return 0, err
	}
	return result.(float64), nil
}

func toFloat(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case int64:
		return float64(v)
	}
	return 0
}

func toInt(val interface{}) int64 {
	switch v := val.(type) {
	case int64:
		return v
	case float64:
		return int64(v)
	}
	return 0
}
