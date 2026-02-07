package solana

import (
	"fmt"
	"time"
)

// =============================================================================
// Parsed entities ready for Neo4j ingestion.
// These are the intermediate representations between raw RPC data and graph nodes.
// =============================================================================

// ParsedTransaction holds all entities extracted from a single Solana transaction.
type ParsedTransaction struct {
	// Core transaction data
	Signature           string
	Slot                uint64
	BlockTime           *time.Time
	Fee                 uint64
	Version             string // "legacy" or "0", "1", etc.
	Err                 interface{}
	ComputeUnitsConsumed *uint64
	RecentBlockhash     string
	LogMessages         []string
	NumRequiredSigs     int
	NumReadonlySigned   int
	NumReadonlyUnsigned int
	ReturnDataProgram   *string
	ReturnData          *string
	ConfirmationStatus  *string
	Memo                *string

	// Fee payer (first signer)
	FeePayer string

	// SOL transfers (derived from pre/post balance deltas)
	SOLTransfers []SOLTransfer

	// Token transfers (derived from pre/post token balance deltas)
	TokenTransfers []TokenTransfer

	// All instructions (top-level + inner)
	Instructions []ParsedInstruction

	// Programs invoked
	Programs []string

	// All wallets referenced in the transaction
	ReferencedWallets []string

	// Address lookup tables used (v0 transactions)
	AddressLookupTables []string

	// Loaded addresses from lookup tables
	LoadedWritableAddresses []string
	LoadedReadonlyAddresses []string

	// Rewards
	Rewards []ParsedReward
}

// SOLTransfer represents a SOL balance change for a wallet.
type SOLTransfer struct {
	Wallet        string
	AmountLamports int64 // positive = received, negative = sent (includes fee for payer)
}

// TokenTransfer represents an SPL token balance change.
type TokenTransfer struct {
	Wallet         string
	Mint           string
	ProgramID      string
	AmountRaw      string  // Raw amount string (to avoid precision loss)
	UIAmount       *float64
	UIAmountString string
	Decimals       int
	Delta          float64 // positive = received, negative = sent
}

// ParsedInstruction is a normalized instruction (works for both top-level and inner).
type ParsedInstruction struct {
	Index          int
	ProgramID      string   // Resolved program address
	Accounts       []string // Resolved account addresses
	Data           string   // Base58-encoded instruction data
	StackHeight    *int
	IsInner        bool
	ParentIndex    int // Index of the parent top-level instruction (for inner instructions)
}

// ParsedReward is a normalized reward entry.
type ParsedReward struct {
	Pubkey      string
	Lamports    int64
	PostBalance uint64
	RewardType  *string
	Commission  *uint8
}

// ParseTransaction converts a raw RPC TransactionResult into a ParsedTransaction.
func ParseTransaction(result *TransactionResult, sigInfo *SignatureInfo) (*ParsedTransaction, error) {
	if result == nil || result.Transaction == nil {
		return nil, fmt.Errorf("nil transaction result")
	}

	tx := result.Transaction
	meta := result.Meta
	msg := tx.Message

	// Build the full account keys list (static + loaded from lookup tables)
	accountKeys := make([]string, len(msg.AccountKeys))
	copy(accountKeys, msg.AccountKeys)
	if meta != nil && meta.LoadedAddresses != nil {
		accountKeys = append(accountKeys, meta.LoadedAddresses.Writable...)
		accountKeys = append(accountKeys, meta.LoadedAddresses.Readonly...)
	}

	parsed := &ParsedTransaction{
		Signature:       tx.Signatures[0],
		Slot:            result.Slot,
		Fee:             meta.Fee,
		RecentBlockhash: msg.RecentBlockhash,
		NumRequiredSigs:     msg.Header.NumRequiredSignatures,
		NumReadonlySigned:   msg.Header.NumReadonlySignedAccounts,
		NumReadonlyUnsigned: msg.Header.NumReadonlyUnsignedAccounts,
	}

	// Block time
	if result.BlockTime != nil {
		t := time.Unix(*result.BlockTime, 0).UTC()
		parsed.BlockTime = &t
	}

	// Version
	switch v := result.Version.(type) {
	case string:
		parsed.Version = v
	case float64:
		parsed.Version = fmt.Sprintf("%d", int(v))
	default:
		parsed.Version = "legacy"
	}

	// Error
	parsed.Err = meta.Err

	// Compute units
	parsed.ComputeUnitsConsumed = meta.ComputeUnitsConsumed

	// Log messages
	parsed.LogMessages = meta.LogMessages

	// Return data
	if meta.ReturnData != nil {
		parsed.ReturnDataProgram = &meta.ReturnData.ProgramID
		if len(meta.ReturnData.Data) > 0 {
			parsed.ReturnData = &meta.ReturnData.Data[0]
		}
	}

	// Carry forward signature-level info
	if sigInfo != nil {
		parsed.ConfirmationStatus = sigInfo.ConfirmationStatus
		parsed.Memo = sigInfo.Memo
	}

	// Fee payer is always the first account key
	if len(accountKeys) > 0 {
		parsed.FeePayer = accountKeys[0]
	}

	// --- SOL Transfers (pre/post balance deltas) ---
	if meta != nil && len(meta.PreBalances) == len(meta.PostBalances) {
		for i := 0; i < len(meta.PreBalances) && i < len(accountKeys); i++ {
			delta := int64(meta.PostBalances[i]) - int64(meta.PreBalances[i])
			if delta != 0 {
				parsed.SOLTransfers = append(parsed.SOLTransfers, SOLTransfer{
					Wallet:         accountKeys[i],
					AmountLamports: delta,
				})
			}
		}
	}

	// --- Token Transfers (pre/post token balance deltas) ---
	if meta != nil {
		parsed.TokenTransfers = computeTokenTransfers(meta.PreTokenBalances, meta.PostTokenBalances, accountKeys)
	}

	// --- Instructions (top-level) ---
	for idx, instr := range msg.Instructions {
		pi := ParsedInstruction{
			Index:       idx,
			Data:        instr.Data,
			StackHeight: instr.StackHeight,
			IsInner:     false,
			ParentIndex: -1,
		}

		// Resolve program ID
		if instr.ProgramIDIndex < len(accountKeys) {
			pi.ProgramID = accountKeys[instr.ProgramIDIndex]
		}

		// Resolve accounts
		for _, accIdx := range instr.Accounts {
			if accIdx < len(accountKeys) {
				pi.Accounts = append(pi.Accounts, accountKeys[accIdx])
			}
		}

		parsed.Instructions = append(parsed.Instructions, pi)
	}

	// --- Inner Instructions ---
	if meta != nil {
		for _, inner := range meta.InnerInstructions {
			for subIdx, instr := range inner.Instructions {
				pi := ParsedInstruction{
					Index:       len(msg.Instructions) + subIdx, // Offset to avoid collision
					Data:        instr.Data,
					StackHeight: instr.StackHeight,
					IsInner:     true,
					ParentIndex: inner.Index,
				}

				if instr.ProgramIDIndex < len(accountKeys) {
					pi.ProgramID = accountKeys[instr.ProgramIDIndex]
				}
				for _, accIdx := range instr.Accounts {
					if accIdx < len(accountKeys) {
						pi.Accounts = append(pi.Accounts, accountKeys[accIdx])
					}
				}

				parsed.Instructions = append(parsed.Instructions, pi)
			}
		}
	}

	// --- Collect unique programs ---
	programSet := make(map[string]bool)
	for _, instr := range parsed.Instructions {
		if instr.ProgramID != "" {
			programSet[instr.ProgramID] = true
		}
	}
	for p := range programSet {
		parsed.Programs = append(parsed.Programs, p)
	}

	// --- Collect all referenced wallets ---
	walletSet := make(map[string]bool)
	for _, key := range accountKeys {
		walletSet[key] = true
	}
	for w := range walletSet {
		parsed.ReferencedWallets = append(parsed.ReferencedWallets, w)
	}

	// --- Address lookup tables ---
	for _, alt := range msg.AddressTableLookups {
		parsed.AddressLookupTables = append(parsed.AddressLookupTables, alt.AccountKey)
	}

	// --- Loaded addresses ---
	if meta != nil && meta.LoadedAddresses != nil {
		parsed.LoadedWritableAddresses = meta.LoadedAddresses.Writable
		parsed.LoadedReadonlyAddresses = meta.LoadedAddresses.Readonly
	}

	// --- Rewards ---
	if meta != nil {
		for _, r := range meta.Rewards {
			parsed.Rewards = append(parsed.Rewards, ParsedReward{
				Pubkey:      r.Pubkey,
				Lamports:    r.Lamports,
				PostBalance: r.PostBalance,
				RewardType:  r.RewardType,
				Commission:  r.Commission,
			})
		}
	}

	return parsed, nil
}

// computeTokenTransfers computes the delta between pre and post token balances.
func computeTokenTransfers(pre, post []TokenBalance, accountKeys []string) []TokenTransfer {
	// Build map of pre-balances keyed by (accountIndex, mint)
	type balKey struct {
		AccountIndex int
		Mint         string
	}
	preMap := make(map[balKey]TokenBalance)
	for _, tb := range pre {
		preMap[balKey{tb.AccountIndex, tb.Mint}] = tb
	}

	var transfers []TokenTransfer

	// Check all post-balances for changes
	seen := make(map[balKey]bool)
	for _, postBal := range post {
		key := balKey{postBal.AccountIndex, postBal.Mint}
		seen[key] = true

		preBal, existed := preMap[key]

		var preAmount, postAmount float64
		if existed && preBal.UITokenAmount.UIAmount != nil {
			preAmount = *preBal.UITokenAmount.UIAmount
		}
		if postBal.UITokenAmount.UIAmount != nil {
			postAmount = *postBal.UITokenAmount.UIAmount
		}

		delta := postAmount - preAmount
		if delta == 0 && existed {
			continue
		}

		wallet := postBal.Owner
		if wallet == "" && postBal.AccountIndex < len(accountKeys) {
			wallet = accountKeys[postBal.AccountIndex]
		}

		transfers = append(transfers, TokenTransfer{
			Wallet:         wallet,
			Mint:           postBal.Mint,
			ProgramID:      postBal.ProgramID,
			AmountRaw:      postBal.UITokenAmount.Amount,
			UIAmount:       postBal.UITokenAmount.UIAmount,
			UIAmountString: postBal.UITokenAmount.UIAmountString,
			Decimals:       postBal.UITokenAmount.Decimals,
			Delta:          delta,
		})
	}

	// Check for accounts that had a pre-balance but no post-balance (fully spent)
	for _, preBal := range pre {
		key := balKey{preBal.AccountIndex, preBal.Mint}
		if seen[key] {
			continue
		}

		var preAmount float64
		if preBal.UITokenAmount.UIAmount != nil {
			preAmount = *preBal.UITokenAmount.UIAmount
		}

		wallet := preBal.Owner
		if wallet == "" && preBal.AccountIndex < len(accountKeys) {
			wallet = accountKeys[preBal.AccountIndex]
		}

		transfers = append(transfers, TokenTransfer{
			Wallet:         wallet,
			Mint:           preBal.Mint,
			ProgramID:      preBal.ProgramID,
			AmountRaw:      "0",
			UIAmount:       nil,
			UIAmountString: "0",
			Decimals:       preBal.UITokenAmount.Decimals,
			Delta:          -preAmount,
		})
	}

	return transfers
}
