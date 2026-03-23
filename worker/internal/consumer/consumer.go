package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/brianbeck/c2graph-worker/internal/config"
	neo4jwriter "github.com/brianbeck/c2graph-worker/internal/neo4j"
	"github.com/brianbeck/c2graph-worker/internal/scoring"
	"github.com/brianbeck/c2graph-worker/internal/solana"
	neo4jdriver "github.com/neo4j/neo4j-go-driver/v5/neo4j"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

// ScanMessage is the message format consumed from the RabbitMQ queue.
type ScanMessage struct {
	JobID        string `json:"job_id"`
	Address      string `json:"address"`
	Depth        int    `json:"depth"`
	CurrentDepth int    `json:"current_depth"`
	RootJobID    string `json:"root_job_id"`
}

// Consumer processes scan requests from RabbitMQ.
type Consumer struct {
	cfg       *config.Config
	solClient *solana.Client
	writer    *neo4jwriter.Writer
	dedup     *neo4jwriter.DedupChecker
	scorer    *scoring.Engine
	conn      *amqp.Connection
	channel   *amqp.Channel

	// Per-job wallet counters: job_id -> count
	jobCounters sync.Map
	// Per-job expected wallet counts: job_id -> expected total
	jobExpected sync.Map
}

// NewConsumer creates a new RabbitMQ consumer.
func NewConsumer(cfg *config.Config, solClient *solana.Client, writer *neo4jwriter.Writer, dedup *neo4jwriter.DedupChecker, scorer *scoring.Engine) *Consumer {
	return &Consumer{
		cfg:       cfg,
		solClient: solClient,
		writer:    writer,
		dedup:     dedup,
		scorer:    scorer,
	}
}

// Start connects to RabbitMQ and begins consuming messages.
func (c *Consumer) Start(ctx context.Context) error {
	var err error

	// Connect to RabbitMQ with retry
	for attempt := 0; attempt < 10; attempt++ {
		c.conn, err = amqp.Dial(c.cfg.RabbitMQURI)
		if err == nil {
			break
		}
		log.Warn().Err(err).Int("attempt", attempt+1).Msg("Failed to connect to RabbitMQ, retrying...")
		select {
		case <-time.After(time.Duration(attempt+1) * 2 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if err != nil {
		return fmt.Errorf("connect to RabbitMQ: %w", err)
	}

	c.channel, err = c.conn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}

	// Declare the queue
	_, err = c.channel.QueueDeclare(
		"scan_requests", // name
		true,            // durable
		false,           // auto-delete
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}

	// Set prefetch count to match concurrency
	if err := c.channel.Qos(c.cfg.Concurrency, 0, false); err != nil {
		return fmt.Errorf("set QoS: %w", err)
	}

	msgs, err := c.channel.Consume(
		"scan_requests", // queue
		"",              // consumer tag (auto-generated)
		false,           // auto-ack (manual ack for reliability)
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	log.Info().Int("concurrency", c.cfg.Concurrency).Msg("Worker started, consuming scan_requests")

	// Process messages with worker pool
	var wg sync.WaitGroup
	for i := 0; i < c.cfg.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case msg, ok := <-msgs:
					if !ok {
						return
					}
					c.processMessage(ctx, workerID, msg)
				case <-ctx.Done():
					return
				}
			}
		}(i)
	}

	wg.Wait()
	return nil
}

// Close cleanly shuts down the consumer.
func (c *Consumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

// PublishScanRequest publishes a new scan message to the queue.
func (c *Consumer) PublishScanRequest(ctx context.Context, msg *ScanMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	return c.channel.PublishWithContext(ctx,
		"",              // exchange (default)
		"scan_requests", // routing key (queue name)
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
}

func (c *Consumer) processMessage(ctx context.Context, workerID int, msg amqp.Delivery) {
	var scanMsg ScanMessage
	if err := json.Unmarshal(msg.Body, &scanMsg); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal scan message")
		msg.Nack(false, false) // Don't requeue malformed messages
		return
	}

	logger := log.With().
		Int("worker", workerID).
		Str("job_id", scanMsg.JobID).
		Str("address", scanMsg.Address).
		Int("depth", scanMsg.CurrentDepth).
		Logger()

	logger.Info().Msg("Processing scan request")

	// Layer 1: Check if wallet is already fresh
	fresh, err := c.dedup.IsWalletFresh(ctx, scanMsg.Address, scanMsg.CurrentDepth)
	if err != nil {
		logger.Error().Err(err).Msg("Freshness check failed")
		msg.Nack(false, true) // Requeue
		return
	}
	if fresh {
		logger.Info().Msg("Wallet data is fresh, skipping")
		c.incrementJobCounter(scanMsg.RootJobID)
		msg.Ack(false)
		return
	}

	// Fetch and process the wallet
	if err := c.processWallet(ctx, &scanMsg); err != nil {
		logger.Error().Err(err).Msg("Failed to process wallet")
		msg.Nack(false, true) // Requeue for retry
		return
	}

	// Score the wallet (bot detection + fraud heuristics)
	if err := c.scorer.ScoreWallet(ctx, scanMsg.Address); err != nil {
		logger.Warn().Err(err).Msg("Failed to score wallet")
	}

	// Mark wallet as scanned
	if err := c.dedup.MarkWalletScanned(ctx, scanMsg.Address, scanMsg.CurrentDepth); err != nil {
		logger.Error().Err(err).Msg("Failed to mark wallet as scanned")
	}

	// Increment job counter
	count := c.incrementJobCounter(scanMsg.RootJobID)

	// If depth > 1, discover and enqueue child wallets
	if scanMsg.CurrentDepth > 1 {
		c.enqueueChildWallets(ctx, &scanMsg, count)
	} else {
		// Leaf node: check if the entire job is complete
		expected := c.getJobExpected(scanMsg.RootJobID)
		if expected > 0 && int(count) >= expected {
			if err := c.dedup.UpdateScanJobProgress(ctx, scanMsg.RootJobID, int(count), "complete"); err != nil {
				logger.Warn().Err(err).Msg("Failed to update job progress")
			}
		} else if expected == 0 && scanMsg.Depth <= 1 {
			// Depth-1 scan: this is the only wallet, job is done
			if err := c.dedup.UpdateScanJobProgress(ctx, scanMsg.RootJobID, int(count), "complete"); err != nil {
				logger.Warn().Err(err).Msg("Failed to update job progress")
			}
		} else {
			if err := c.dedup.UpdateScanJobProgress(ctx, scanMsg.RootJobID, int(count), "processing"); err != nil {
				logger.Warn().Err(err).Msg("Failed to update job progress")
			}
		}
	}

	msg.Ack(false)
	logger.Info().Int64("job_total", count).Msg("Wallet processed successfully")
}

func (c *Consumer) processWallet(ctx context.Context, scanMsg *ScanMessage) error {
	// Step 1: Fetch all signatures for this wallet
	sigs, err := c.solClient.GetAllSignaturesForAddress(ctx, scanMsg.Address)
	if err != nil {
		return fmt.Errorf("fetch signatures: %w", err)
	}
	log.Info().Str("address", scanMsg.Address).Int("total_sigs", len(sigs)).Msg("Fetched signatures")

	// Step 2: Filter out signatures we already have (Layer 2 dedup)
	sigStrings := make([]string, len(sigs))
	sigMap := make(map[string]*solana.SignatureInfo, len(sigs))
	for i, sig := range sigs {
		sigStrings[i] = sig.Signature
		s := sigs[i] // Copy to avoid reference issues
		sigMap[sig.Signature] = &s
	}

	newSigs, err := c.dedup.FilterNewSignatures(ctx, sigStrings)
	if err != nil {
		return fmt.Errorf("filter signatures: %w", err)
	}
	log.Info().Str("address", scanMsg.Address).Int("new_sigs", len(newSigs)).Int("skipped", len(sigs)-len(newSigs)).Msg("Filtered known signatures")

	// Step 3: Fetch full transaction details for new signatures and write to Neo4j
	var batch []*solana.ParsedTransaction
	for _, sig := range newSigs {
		txResult, err := c.solClient.GetTransaction(ctx, sig)
		if err != nil {
			log.Warn().Err(err).Str("signature", sig).Msg("Failed to fetch transaction, skipping")
			continue
		}
		if txResult == nil {
			continue
		}

		parsed, err := solana.ParseTransaction(txResult, sigMap[sig])
		if err != nil {
			log.Warn().Err(err).Str("signature", sig).Msg("Failed to parse transaction, skipping")
			continue
		}

		batch = append(batch, parsed)

		// Write batch when it reaches configured size
		if len(batch) >= c.cfg.BatchSize {
			if err := c.writer.WriteBatch(ctx, batch); err != nil {
				return fmt.Errorf("write batch: %w", err)
			}
			log.Debug().Int("batch_size", len(batch)).Msg("Wrote batch to Neo4j")
			batch = batch[:0]
		}
	}

	// Write remaining batch
	if len(batch) > 0 {
		if err := c.writer.WriteBatch(ctx, batch); err != nil {
			return fmt.Errorf("write final batch: %w", err)
		}
	}

	// Step 4: Fetch account info and update wallet node
	acctInfo, err := c.solClient.GetAccountInfo(ctx, scanMsg.Address)
	if err != nil {
		log.Warn().Err(err).Str("address", scanMsg.Address).Msg("Failed to fetch account info")
	} else if acctInfo != nil && acctInfo.Value != nil {
		if err := c.updateWalletAccountInfo(ctx, scanMsg.Address, acctInfo.Value); err != nil {
			log.Warn().Err(err).Str("address", scanMsg.Address).Msg("Failed to update wallet account info")
		}
	}

	// Step 5: Fetch token holdings
	tokenAccounts, err := c.solClient.GetTokenAccountsByOwner(ctx, scanMsg.Address)
	if err != nil {
		log.Warn().Err(err).Str("address", scanMsg.Address).Msg("Failed to fetch token accounts")
	} else if tokenAccounts != nil {
		if err := c.writeTokenHoldings(ctx, scanMsg.Address, tokenAccounts); err != nil {
			log.Warn().Err(err).Str("address", scanMsg.Address).Msg("Failed to write token holdings")
		}
	}

	return nil
}

func (c *Consumer) updateWalletAccountInfo(ctx context.Context, address string, info *solana.AccountInfoValue) error {
	session := c.writer.GetDriver().NewSession(ctx, neo4jdriver.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4jdriver.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4jdriver.ManagedTransaction) (interface{}, error) {
		query := `
			MERGE (w:Wallet {address: $address})
			SET w.lamports = $lamports,
			    w.executable = $executable,
			    w.owner = $owner,
			    w.rent_epoch = $rent_epoch,
			    w.space = $space
		`
		_, err := tx.Run(ctx, query, map[string]interface{}{
			"address":    address,
			"lamports":   int64(info.Lamports),
			"executable": info.Executable,
			"owner":      info.Owner,
			"rent_epoch": int64(info.RentEpoch),
			"space":      int64(info.Space),
		})
		return nil, err
	})
	return err
}

func (c *Consumer) writeTokenHoldings(ctx context.Context, address string, accounts *solana.TokenAccountsResult) error {
	session := c.writer.GetDriver().NewSession(ctx, neo4jdriver.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4jdriver.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4jdriver.ManagedTransaction) (interface{}, error) {
		for _, acct := range accounts.Value {
			info := acct.Account.Data.Parsed.Info
			// Merge token node
			tokenQuery := `
				MERGE (t:Token {mint_address: $mint})
				ON CREATE SET t.decimals = $decimals
			`
			if _, err := tx.Run(ctx, tokenQuery, map[string]interface{}{
				"mint":     info.Mint,
				"decimals": info.TokenAmount.Decimals,
			}); err != nil {
				return nil, err
			}

			// Create/update HOLDS relationship
			holdsQuery := `
				MATCH (w:Wallet {address: $address}), (t:Token {mint_address: $mint})
				MERGE (w)-[h:HOLDS]->(t)
				SET h.balance = $balance,
				    h.ui_balance = $ui_balance,
				    h.decimals = $decimals,
				    h.is_native = $is_native,
				    h.state = $state
			`
			if _, err := tx.Run(ctx, holdsQuery, map[string]interface{}{
				"address":    address,
				"mint":       info.Mint,
				"balance":    info.TokenAmount.Amount,
				"ui_balance": info.TokenAmount.UIAmountString,
				"decimals":   info.TokenAmount.Decimals,
				"is_native":  info.IsNative,
				"state":      info.State,
			}); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return err
}

func (c *Consumer) enqueueChildWallets(ctx context.Context, scanMsg *ScanMessage, currentJobCount int64) {
	// Check cap
	if int(currentJobCount) >= c.cfg.MaxWalletsPerScan {
		log.Info().Str("job_id", scanMsg.RootJobID).Int("cap", c.cfg.MaxWalletsPerScan).Msg("Wallet cap reached, stopping traversal")
		c.dedup.UpdateScanJobProgress(ctx, scanMsg.RootJobID, int(currentJobCount), "capped")
		return
	}

	// Discover connected wallets
	childWallets, err := c.dedup.GetDiscoveredWallets(ctx, scanMsg.Address)
	if err != nil {
		log.Error().Err(err).Str("address", scanMsg.Address).Msg("Failed to discover child wallets")
		// No children to enqueue — if this is the root, the job is complete
		c.dedup.UpdateScanJobProgress(ctx, scanMsg.RootJobID, int(currentJobCount), "complete")
		return
	}

	remaining := c.cfg.MaxWalletsPerScan - int(currentJobCount)
	enqueued := 0

	for _, childAddr := range childWallets {
		if enqueued >= remaining {
			break
		}

		// Check if child is already fresh
		fresh, err := c.dedup.IsWalletFresh(ctx, childAddr, scanMsg.CurrentDepth-1)
		if err != nil {
			log.Warn().Err(err).Str("address", childAddr).Msg("Freshness check failed for child wallet")
			continue
		}
		if fresh {
			continue
		}

		childMsg := &ScanMessage{
			JobID:        fmt.Sprintf("%s-%s", scanMsg.RootJobID, childAddr[:8]),
			Address:      childAddr,
			Depth:        scanMsg.Depth,
			CurrentDepth: scanMsg.CurrentDepth - 1,
			RootJobID:    scanMsg.RootJobID,
		}

		if err := c.PublishScanRequest(ctx, childMsg); err != nil {
			log.Error().Err(err).Str("address", childAddr).Msg("Failed to enqueue child wallet")
			continue
		}
		enqueued++
	}

	log.Info().Str("address", scanMsg.Address).Int("children", enqueued).Msg("Enqueued child wallets")

	if enqueued == 0 {
		// No children enqueued — job is complete
		c.dedup.UpdateScanJobProgress(ctx, scanMsg.RootJobID, int(currentJobCount), "complete")
	} else {
		// Set expected count so leaf nodes know when the job is done
		// Expected = wallets already processed + children just enqueued
		expected := int(currentJobCount) + enqueued
		c.setJobExpected(scanMsg.RootJobID, expected)
		c.dedup.SetScanJobExpected(ctx, scanMsg.RootJobID, expected)
		c.dedup.UpdateScanJobProgress(ctx, scanMsg.RootJobID, int(currentJobCount), "processing")
	}
}

func (c *Consumer) incrementJobCounter(jobID string) int64 {
	val, _ := c.jobCounters.LoadOrStore(jobID, new(atomic.Int64))
	counter := val.(*atomic.Int64)
	return counter.Add(1)
}

func (c *Consumer) setJobExpected(jobID string, expected int) {
	c.jobExpected.Store(jobID, expected)
}

func (c *Consumer) getJobExpected(jobID string) int {
	val, ok := c.jobExpected.Load(jobID)
	if !ok {
		return 0
	}
	return val.(int)
}
