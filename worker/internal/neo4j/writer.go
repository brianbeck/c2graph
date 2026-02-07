package neo4j

import (
	"context"
	"fmt"
	"time"

	"github.com/brianbeck/sentinel-worker/internal/solana"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/rs/zerolog/log"
)

// Writer handles writing parsed Solana data to Neo4j.
type Writer struct {
	driver neo4j.DriverWithContext
}

// NewWriter creates a new Neo4j writer.
func NewWriter(driver neo4j.DriverWithContext) *Writer {
	return &Writer{driver: driver}
}

// GetDriver returns the underlying Neo4j driver (used by consumer for direct queries).
func (w *Writer) GetDriver() neo4j.DriverWithContext {
	return w.driver
}

// WriteParsedTransaction writes a fully parsed transaction and all its entities to Neo4j.
func (w *Writer) WriteParsedTransaction(ctx context.Context, ptx *solana.ParsedTransaction) error {
	session := w.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 1. MERGE the Transaction node
		if err := mergeTransaction(ctx, tx, ptx); err != nil {
			return nil, fmt.Errorf("merge transaction: %w", err)
		}

		// 2. MERGE the fee payer wallet + INITIATED relationship
		if ptx.FeePayer != "" {
			if err := mergeWalletAndRelationship(ctx, tx, ptx.FeePayer, ptx.Signature, "INITIATED", nil); err != nil {
				return nil, fmt.Errorf("merge fee payer: %w", err)
			}
		}

		// 3. Write SOL transfers
		for _, transfer := range ptx.SOLTransfers {
			relType := "TRANSFERRED_SOL"
			direction := "receive"
			if transfer.AmountLamports < 0 {
				direction = "send"
			}
			props := map[string]interface{}{
				"amount_lamports": transfer.AmountLamports,
				"direction":       direction,
			}
			if err := mergeWalletAndRelationship(ctx, tx, transfer.Wallet, ptx.Signature, relType, props); err != nil {
				return nil, fmt.Errorf("merge SOL transfer: %w", err)
			}
		}

		// 4. Write token transfers
		for _, transfer := range ptx.TokenTransfers {
			direction := "receive"
			if transfer.Delta < 0 {
				direction = "send"
			}
			props := map[string]interface{}{
				"amount":    transfer.AmountRaw,
				"ui_amount": transfer.UIAmountString,
				"decimals":  transfer.Decimals,
				"direction": direction,
				"mint":      transfer.Mint,
			}
			if err := mergeWalletAndRelationship(ctx, tx, transfer.Wallet, ptx.Signature, "TRANSFERRED_TOKEN", props); err != nil {
				return nil, fmt.Errorf("merge token transfer: %w", err)
			}

			// MERGE the Token node and INVOLVES_TOKEN relationship
			if err := mergeToken(ctx, tx, transfer.Mint, transfer.Decimals, transfer.ProgramID); err != nil {
				return nil, fmt.Errorf("merge token: %w", err)
			}
			if err := mergeTransactionTokenRelationship(ctx, tx, ptx.Signature, transfer.Mint); err != nil {
				return nil, fmt.Errorf("merge tx-token rel: %w", err)
			}
		}

		// 5. Write programs + EXECUTED_BY relationships via instructions
		for _, prog := range ptx.Programs {
			if err := mergeProgram(ctx, tx, prog); err != nil {
				return nil, fmt.Errorf("merge program: %w", err)
			}
		}

		// 6. Write instructions
		for _, instr := range ptx.Instructions {
			if err := mergeInstruction(ctx, tx, ptx.Signature, &instr); err != nil {
				return nil, fmt.Errorf("merge instruction: %w", err)
			}
		}

		// 7. Write address lookup tables
		for _, alt := range ptx.AddressLookupTables {
			if err := mergeAddressLookupTable(ctx, tx, ptx.Signature, alt); err != nil {
				return nil, fmt.Errorf("merge ALT: %w", err)
			}
		}

		// 8. Write loaded addresses
		for _, addr := range ptx.LoadedWritableAddresses {
			if err := mergeLoadedAddress(ctx, tx, ptx.Signature, addr, true); err != nil {
				return nil, fmt.Errorf("merge loaded writable: %w", err)
			}
		}
		for _, addr := range ptx.LoadedReadonlyAddresses {
			if err := mergeLoadedAddress(ctx, tx, ptx.Signature, addr, false); err != nil {
				return nil, fmt.Errorf("merge loaded readonly: %w", err)
			}
		}

		// 9. Write rewards
		for _, reward := range ptx.Rewards {
			if err := mergeReward(ctx, tx, ptx.Signature, &reward); err != nil {
				return nil, fmt.Errorf("merge reward: %w", err)
			}
		}

		return nil, nil
	})

	if err != nil {
		log.Error().Err(err).Str("signature", ptx.Signature).Msg("Failed to write transaction to Neo4j")
	}
	return err
}

// WriteBatch writes a batch of parsed transactions to Neo4j.
func (w *Writer) WriteBatch(ctx context.Context, batch []*solana.ParsedTransaction) error {
	for _, ptx := range batch {
		if err := w.WriteParsedTransaction(ctx, ptx); err != nil {
			return err
		}
	}
	return nil
}

// --- Individual MERGE helpers ---

func mergeTransaction(ctx context.Context, tx neo4j.ManagedTransaction, ptx *solana.ParsedTransaction) error {
	query := `
		MERGE (t:Transaction {signature: $signature})
		SET t.slot = $slot,
		    t.fee = $fee,
		    t.version = $version,
		    t.recent_blockhash = $recent_blockhash,
		    t.num_required_signatures = $num_required_sigs,
		    t.num_readonly_signed = $num_readonly_signed,
		    t.num_readonly_unsigned = $num_readonly_unsigned,
		    t.compute_units_consumed = $compute_units,
		    t.log_messages = $log_messages,
		    t.return_data_program = $return_data_program,
		    t.return_data = $return_data,
		    t.confirmation_status = $confirmation_status,
		    t.memo = $memo
		SET t.block_time = CASE WHEN $block_time IS NOT NULL THEN datetime($block_time) ELSE NULL END
		SET t.err = CASE WHEN $err IS NOT NULL THEN $err ELSE NULL END
	`

	var blockTimeStr *string
	if ptx.BlockTime != nil {
		s := ptx.BlockTime.Format(time.RFC3339)
		blockTimeStr = &s
	}

	var errStr *string
	if ptx.Err != nil {
		s := fmt.Sprintf("%v", ptx.Err)
		errStr = &s
	}

	var computeUnits *int64
	if ptx.ComputeUnitsConsumed != nil {
		cu := int64(*ptx.ComputeUnitsConsumed)
		computeUnits = &cu
	}

	params := map[string]interface{}{
		"signature":            ptx.Signature,
		"slot":                 int64(ptx.Slot),
		"fee":                  int64(ptx.Fee),
		"version":              ptx.Version,
		"recent_blockhash":     ptx.RecentBlockhash,
		"num_required_sigs":    ptx.NumRequiredSigs,
		"num_readonly_signed":  ptx.NumReadonlySigned,
		"num_readonly_unsigned": ptx.NumReadonlyUnsigned,
		"compute_units":        computeUnits,
		"log_messages":         ptx.LogMessages,
		"return_data_program":  ptx.ReturnDataProgram,
		"return_data":          ptx.ReturnData,
		"confirmation_status":  ptx.ConfirmationStatus,
		"memo":                 ptx.Memo,
		"block_time":           blockTimeStr,
		"err":                  errStr,
	}

	_, err := tx.Run(ctx, query, params)
	return err
}

func mergeWalletAndRelationship(ctx context.Context, tx neo4j.ManagedTransaction, address, txSig, relType string, props map[string]interface{}) error {
	// First ensure the wallet exists
	walletQuery := `
		MERGE (w:Wallet {address: $address})
		ON CREATE SET w.first_seen = datetime(), w.tx_count = 0, w.risk_score = 0.0
		SET w.last_seen = datetime(),
		    w.tx_count = COALESCE(w.tx_count, 0) + 1
	`
	if _, err := tx.Run(ctx, walletQuery, map[string]interface{}{"address": address}); err != nil {
		return err
	}

	// Create the relationship from/to the transaction
	var relQuery string
	params := map[string]interface{}{
		"address": address,
		"tx_sig":  txSig,
	}

	switch relType {
	case "INITIATED":
		relQuery = `
			MATCH (w:Wallet {address: $address}), (t:Transaction {signature: $tx_sig})
			MERGE (w)-[:INITIATED]->(t)
		`
	case "TRANSFERRED_SOL":
		relQuery = `
			MATCH (w:Wallet {address: $address}), (t:Transaction {signature: $tx_sig})
			MERGE (t)-[r:TRANSFERRED_SOL]->(w)
			SET r.amount_lamports = $amount_lamports, r.direction = $direction
		`
		for k, v := range props {
			params[k] = v
		}
	case "TRANSFERRED_TOKEN":
		relQuery = `
			MATCH (w:Wallet {address: $address}), (t:Transaction {signature: $tx_sig})
			MERGE (t)-[r:TRANSFERRED_TOKEN {mint: $mint}]->(w)
			SET r.amount = $amount, r.ui_amount = $ui_amount,
			    r.decimals = $decimals, r.direction = $direction
		`
		for k, v := range props {
			params[k] = v
		}
	default:
		return fmt.Errorf("unknown relationship type: %s", relType)
	}

	_, err := tx.Run(ctx, relQuery, params)
	return err
}

func mergeToken(ctx context.Context, tx neo4j.ManagedTransaction, mint string, decimals int, programID string) error {
	query := `
		MERGE (t:Token {mint_address: $mint})
		ON CREATE SET t.decimals = $decimals, t.program_id = $program_id
	`
	_, err := tx.Run(ctx, query, map[string]interface{}{
		"mint":       mint,
		"decimals":   decimals,
		"program_id": programID,
	})
	return err
}

func mergeTransactionTokenRelationship(ctx context.Context, tx neo4j.ManagedTransaction, txSig, mint string) error {
	query := `
		MATCH (t:Transaction {signature: $tx_sig}), (tok:Token {mint_address: $mint})
		MERGE (t)-[:INVOLVES_TOKEN]->(tok)
	`
	_, err := tx.Run(ctx, query, map[string]interface{}{
		"tx_sig": txSig,
		"mint":   mint,
	})
	return err
}

func mergeProgram(ctx context.Context, tx neo4j.ManagedTransaction, programID string) error {
	query := `
		MERGE (p:Program {program_id: $program_id})
	`
	_, err := tx.Run(ctx, query, map[string]interface{}{
		"program_id": programID,
	})
	return err
}

func mergeInstruction(ctx context.Context, tx neo4j.ManagedTransaction, txSig string, instr *solana.ParsedInstruction) error {
	query := `
		MATCH (t:Transaction {signature: $tx_sig})
		CREATE (i:Instruction {
			index: $index,
			data: $data,
			stack_height: $stack_height,
			is_inner: $is_inner,
			parent_index: $parent_index
		})
		CREATE (t)-[:CONTAINS_INSTRUCTION {order: $index}]->(i)
		WITH i
		MATCH (p:Program {program_id: $program_id})
		MERGE (i)-[:EXECUTED_BY]->(p)
	`

	var stackHeight interface{}
	if instr.StackHeight != nil {
		stackHeight = *instr.StackHeight
	}

	_, err := tx.Run(ctx, query, map[string]interface{}{
		"tx_sig":       txSig,
		"index":        instr.Index,
		"data":         instr.Data,
		"stack_height": stackHeight,
		"is_inner":     instr.IsInner,
		"parent_index": instr.ParentIndex,
		"program_id":   instr.ProgramID,
	})
	return err
}

func mergeAddressLookupTable(ctx context.Context, tx neo4j.ManagedTransaction, txSig, altAddress string) error {
	query := `
		MERGE (a:AddressLookupTable {address: $alt_address})
		WITH a
		MATCH (t:Transaction {signature: $tx_sig})
		MERGE (t)-[:USED_LOOKUP_TABLE]->(a)
	`
	_, err := tx.Run(ctx, query, map[string]interface{}{
		"tx_sig":      txSig,
		"alt_address": altAddress,
	})
	return err
}

func mergeLoadedAddress(ctx context.Context, tx neo4j.ManagedTransaction, txSig, address string, writable bool) error {
	query := `
		MERGE (w:Wallet {address: $address})
		ON CREATE SET w.first_seen = datetime(), w.tx_count = 0, w.risk_score = 0.0
		WITH w
		MATCH (t:Transaction {signature: $tx_sig})
		MERGE (t)-[:LOADED_ADDRESS {writable: $writable}]->(w)
	`
	_, err := tx.Run(ctx, query, map[string]interface{}{
		"tx_sig":   txSig,
		"address":  address,
		"writable": writable,
	})
	return err
}

func mergeReward(ctx context.Context, tx neo4j.ManagedTransaction, txSig string, reward *solana.ParsedReward) error {
	query := `
		MERGE (w:Wallet {address: $pubkey})
		ON CREATE SET w.first_seen = datetime(), w.tx_count = 0, w.risk_score = 0.0
		WITH w
		MATCH (t:Transaction {signature: $tx_sig})
		MERGE (w)-[r:REWARDED_IN]->(t)
		SET r.lamports = $lamports, r.post_balance = $post_balance,
		    r.reward_type = $reward_type, r.commission = $commission
	`

	var commission interface{}
	if reward.Commission != nil {
		commission = int(*reward.Commission)
	}

	_, err := tx.Run(ctx, query, map[string]interface{}{
		"tx_sig":       txSig,
		"pubkey":       reward.Pubkey,
		"lamports":     reward.Lamports,
		"post_balance": int64(reward.PostBalance),
		"reward_type":  reward.RewardType,
		"commission":   commission,
	})
	return err
}
