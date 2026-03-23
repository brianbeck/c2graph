package solana

import (
	"testing"
	"time"
)

func ptr[T any](v T) *T { return &v }

func makeBasicTransactionResult() *TransactionResult {
	blockTime := int64(1700000000)
	computeUnits := uint64(5000)

	return &TransactionResult{
		Slot:      123456,
		BlockTime: &blockTime,
		Version:   "legacy",
		Transaction: &Transaction{
			Signatures: []string{"sig123abc"},
			Message: Message{
				AccountKeys:     []string{"walletA", "walletB", "programX"},
				RecentBlockhash: "blockhash123",
				Header: Header{
					NumRequiredSignatures:       1,
					NumReadonlySignedAccounts:   0,
					NumReadonlyUnsignedAccounts: 1,
				},
				Instructions: []Instruction{
					{
						ProgramIDIndex: 2,
						Accounts:       []int{0, 1},
						Data:           "instructionData",
					},
				},
			},
		},
		Meta: &Meta{
			Fee:                  5000,
			PreBalances:          []uint64{1000000000, 500000000, 0},
			PostBalances:         []uint64{999995000, 500005000, 0},
			ComputeUnitsConsumed: &computeUnits,
			LogMessages:          []string{"Program X invoked", "Program X success"},
		},
	}
}

func TestParseTransaction_NilInput(t *testing.T) {
	_, err := ParseTransaction(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil input")
	}
}

func TestParseTransaction_NilTransaction(t *testing.T) {
	result := &TransactionResult{Transaction: nil}
	_, err := ParseTransaction(result, nil)
	if err == nil {
		t.Fatal("expected error for nil transaction")
	}
}

func TestParseTransaction_BasicFields(t *testing.T) {
	result := makeBasicTransactionResult()
	parsed, err := ParseTransaction(result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed.Signature != "sig123abc" {
		t.Errorf("Signature = %q, want %q", parsed.Signature, "sig123abc")
	}
	if parsed.Slot != 123456 {
		t.Errorf("Slot = %d, want %d", parsed.Slot, 123456)
	}
	if parsed.Fee != 5000 {
		t.Errorf("Fee = %d, want %d", parsed.Fee, 5000)
	}
	if parsed.Version != "legacy" {
		t.Errorf("Version = %q, want %q", parsed.Version, "legacy")
	}
	if parsed.FeePayer != "walletA" {
		t.Errorf("FeePayer = %q, want %q", parsed.FeePayer, "walletA")
	}
	if parsed.RecentBlockhash != "blockhash123" {
		t.Errorf("RecentBlockhash = %q", parsed.RecentBlockhash)
	}
	if parsed.NumRequiredSigs != 1 {
		t.Errorf("NumRequiredSigs = %d", parsed.NumRequiredSigs)
	}
}

func TestParseTransaction_BlockTime(t *testing.T) {
	result := makeBasicTransactionResult()
	parsed, err := ParseTransaction(result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed.BlockTime == nil {
		t.Fatal("BlockTime should not be nil")
	}
	expected := time.Unix(1700000000, 0).UTC()
	if !parsed.BlockTime.Equal(expected) {
		t.Errorf("BlockTime = %v, want %v", parsed.BlockTime, expected)
	}
}

func TestParseTransaction_NilBlockTime(t *testing.T) {
	result := makeBasicTransactionResult()
	result.BlockTime = nil
	parsed, err := ParseTransaction(result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.BlockTime != nil {
		t.Errorf("BlockTime should be nil, got %v", parsed.BlockTime)
	}
}

func TestParseTransaction_SOLTransfers(t *testing.T) {
	result := makeBasicTransactionResult()
	parsed, err := ParseTransaction(result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// walletA: 1000000000 -> 999995000 = -5000 (fee payer, lost 5000)
	// walletB: 500000000 -> 500005000 = +5000
	// programX: 0 -> 0 = no change
	if len(parsed.SOLTransfers) != 2 {
		t.Fatalf("SOLTransfers count = %d, want 2", len(parsed.SOLTransfers))
	}

	transferMap := make(map[string]int64)
	for _, st := range parsed.SOLTransfers {
		transferMap[st.Wallet] = st.AmountLamports
	}

	if transferMap["walletA"] != -5000 {
		t.Errorf("walletA delta = %d, want -5000", transferMap["walletA"])
	}
	if transferMap["walletB"] != 5000 {
		t.Errorf("walletB delta = %d, want 5000", transferMap["walletB"])
	}
}

func TestParseTransaction_Instructions(t *testing.T) {
	result := makeBasicTransactionResult()
	parsed, err := ParseTransaction(result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(parsed.Instructions) != 1 {
		t.Fatalf("Instructions count = %d, want 1", len(parsed.Instructions))
	}

	instr := parsed.Instructions[0]
	if instr.ProgramID != "programX" {
		t.Errorf("ProgramID = %q, want %q", instr.ProgramID, "programX")
	}
	if instr.Data != "instructionData" {
		t.Errorf("Data = %q", instr.Data)
	}
	if instr.IsInner {
		t.Error("IsInner should be false for top-level instruction")
	}
	if len(instr.Accounts) != 2 {
		t.Errorf("Accounts count = %d, want 2", len(instr.Accounts))
	}
	if instr.Accounts[0] != "walletA" || instr.Accounts[1] != "walletB" {
		t.Errorf("Accounts = %v, want [walletA, walletB]", instr.Accounts)
	}
}

func TestParseTransaction_InnerInstructions(t *testing.T) {
	result := makeBasicTransactionResult()
	sh := 2
	result.Meta.InnerInstructions = []InnerInstruction{
		{
			Index: 0,
			Instructions: []InstructionDetail{
				{
					ProgramIDIndex: 2,
					Accounts:       []int{0},
					Data:           "innerData",
					StackHeight:    &sh,
				},
			},
		},
	}

	parsed, err := ParseTransaction(result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 1 top-level + 1 inner
	if len(parsed.Instructions) != 2 {
		t.Fatalf("Instructions count = %d, want 2", len(parsed.Instructions))
	}

	inner := parsed.Instructions[1]
	if !inner.IsInner {
		t.Error("second instruction should be inner")
	}
	if inner.ParentIndex != 0 {
		t.Errorf("ParentIndex = %d, want 0", inner.ParentIndex)
	}
	if inner.ProgramID != "programX" {
		t.Errorf("ProgramID = %q", inner.ProgramID)
	}
}

func TestParseTransaction_Programs(t *testing.T) {
	result := makeBasicTransactionResult()
	parsed, err := ParseTransaction(result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(parsed.Programs) != 1 {
		t.Fatalf("Programs count = %d, want 1", len(parsed.Programs))
	}
	if parsed.Programs[0] != "programX" {
		t.Errorf("Programs[0] = %q, want %q", parsed.Programs[0], "programX")
	}
}

func TestParseTransaction_ReferencedWallets(t *testing.T) {
	result := makeBasicTransactionResult()
	parsed, err := ParseTransaction(result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	walletSet := make(map[string]bool)
	for _, w := range parsed.ReferencedWallets {
		walletSet[w] = true
	}

	for _, expected := range []string{"walletA", "walletB", "programX"} {
		if !walletSet[expected] {
			t.Errorf("missing expected wallet %q", expected)
		}
	}
}

func TestParseTransaction_VersionFormats(t *testing.T) {
	tests := []struct {
		name    string
		version interface{}
		want    string
	}{
		{"legacy string", "legacy", "legacy"},
		{"v0 float", float64(0), "0"},
		{"v1 float", float64(1), "1"},
		{"nil", nil, "legacy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeBasicTransactionResult()
			result.Version = tt.version
			parsed, err := ParseTransaction(result, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if parsed.Version != tt.want {
				t.Errorf("Version = %q, want %q", parsed.Version, tt.want)
			}
		})
	}
}

func TestParseTransaction_ComputeUnits(t *testing.T) {
	result := makeBasicTransactionResult()
	parsed, err := ParseTransaction(result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed.ComputeUnitsConsumed == nil {
		t.Fatal("ComputeUnitsConsumed should not be nil")
	}
	if *parsed.ComputeUnitsConsumed != 5000 {
		t.Errorf("ComputeUnitsConsumed = %d, want 5000", *parsed.ComputeUnitsConsumed)
	}
}

func TestParseTransaction_LogMessages(t *testing.T) {
	result := makeBasicTransactionResult()
	parsed, err := ParseTransaction(result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(parsed.LogMessages) != 2 {
		t.Errorf("LogMessages count = %d, want 2", len(parsed.LogMessages))
	}
}

func TestParseTransaction_SigInfo(t *testing.T) {
	result := makeBasicTransactionResult()
	sigInfo := &SignatureInfo{
		ConfirmationStatus: ptr("finalized"),
		Memo:               ptr("test memo"),
	}

	parsed, err := ParseTransaction(result, sigInfo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed.ConfirmationStatus == nil || *parsed.ConfirmationStatus != "finalized" {
		t.Errorf("ConfirmationStatus = %v", parsed.ConfirmationStatus)
	}
	if parsed.Memo == nil || *parsed.Memo != "test memo" {
		t.Errorf("Memo = %v", parsed.Memo)
	}
}

func TestParseTransaction_LoadedAddresses(t *testing.T) {
	result := makeBasicTransactionResult()
	result.Meta.LoadedAddresses = &LoadedAddresses{
		Writable: []string{"loadedW1", "loadedW2"},
		Readonly: []string{"loadedR1"},
	}

	parsed, err := ParseTransaction(result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(parsed.LoadedWritableAddresses) != 2 {
		t.Errorf("LoadedWritableAddresses count = %d, want 2", len(parsed.LoadedWritableAddresses))
	}
	if len(parsed.LoadedReadonlyAddresses) != 1 {
		t.Errorf("LoadedReadonlyAddresses count = %d, want 1", len(parsed.LoadedReadonlyAddresses))
	}

	// Loaded addresses should also be in ReferencedWallets (appended to accountKeys)
	walletSet := make(map[string]bool)
	for _, w := range parsed.ReferencedWallets {
		walletSet[w] = true
	}
	if !walletSet["loadedW1"] {
		t.Error("loadedW1 should be in ReferencedWallets")
	}
	if !walletSet["loadedR1"] {
		t.Error("loadedR1 should be in ReferencedWallets")
	}
}

func TestParseTransaction_AddressLookupTables(t *testing.T) {
	result := makeBasicTransactionResult()
	result.Transaction.Message.AddressTableLookups = []AddressTableLookup{
		{
			AccountKey:      "altAddress1",
			WritableIndexes: []int{0, 1},
			ReadonlyIndexes: []int{2},
		},
	}

	parsed, err := ParseTransaction(result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(parsed.AddressLookupTables) != 1 {
		t.Fatalf("AddressLookupTables count = %d, want 1", len(parsed.AddressLookupTables))
	}
	if parsed.AddressLookupTables[0] != "altAddress1" {
		t.Errorf("AddressLookupTables[0] = %q", parsed.AddressLookupTables[0])
	}
}

func TestParseTransaction_Rewards(t *testing.T) {
	result := makeBasicTransactionResult()
	rewardType := "rent"
	commission := uint8(5)
	result.Meta.Rewards = []Reward{
		{
			Pubkey:      "rewardWallet",
			Lamports:    1000,
			PostBalance: 50000,
			RewardType:  &rewardType,
			Commission:  &commission,
		},
	}

	parsed, err := ParseTransaction(result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(parsed.Rewards) != 1 {
		t.Fatalf("Rewards count = %d, want 1", len(parsed.Rewards))
	}

	r := parsed.Rewards[0]
	if r.Pubkey != "rewardWallet" {
		t.Errorf("Pubkey = %q", r.Pubkey)
	}
	if r.Lamports != 1000 {
		t.Errorf("Lamports = %d", r.Lamports)
	}
	if r.PostBalance != 50000 {
		t.Errorf("PostBalance = %d", r.PostBalance)
	}
	if r.RewardType == nil || *r.RewardType != "rent" {
		t.Errorf("RewardType = %v", r.RewardType)
	}
	if r.Commission == nil || *r.Commission != 5 {
		t.Errorf("Commission = %v", r.Commission)
	}
}

func TestParseTransaction_ReturnData(t *testing.T) {
	result := makeBasicTransactionResult()
	result.Meta.ReturnData = &ReturnData{
		ProgramID: "progReturnData",
		Data:      []string{"base64data", "base64"},
	}

	parsed, err := ParseTransaction(result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed.ReturnDataProgram == nil || *parsed.ReturnDataProgram != "progReturnData" {
		t.Errorf("ReturnDataProgram = %v", parsed.ReturnDataProgram)
	}
	if parsed.ReturnData == nil || *parsed.ReturnData != "base64data" {
		t.Errorf("ReturnData = %v", parsed.ReturnData)
	}
}

func TestParseTransaction_Error(t *testing.T) {
	result := makeBasicTransactionResult()
	result.Meta.Err = map[string]interface{}{"InstructionError": []interface{}{0, "Custom"}}

	parsed, err := ParseTransaction(result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed.Err == nil {
		t.Error("Err should not be nil")
	}
}

// --- computeTokenTransfers tests ---

func TestComputeTokenTransfers_BasicDelta(t *testing.T) {
	pre := []TokenBalance{
		{
			AccountIndex: 0,
			Mint:         "mintA",
			Owner:        "wallet1",
			ProgramID:    "TokenProg",
			UITokenAmount: UITokenAmount{
				Amount:         "1000",
				Decimals:       6,
				UIAmount:       ptr(0.001),
				UIAmountString: "0.001",
			},
		},
	}
	post := []TokenBalance{
		{
			AccountIndex: 0,
			Mint:         "mintA",
			Owner:        "wallet1",
			ProgramID:    "TokenProg",
			UITokenAmount: UITokenAmount{
				Amount:         "2000",
				Decimals:       6,
				UIAmount:       ptr(0.002),
				UIAmountString: "0.002",
			},
		},
	}

	transfers := computeTokenTransfers(pre, post, []string{"wallet1"})
	if len(transfers) != 1 {
		t.Fatalf("transfers count = %d, want 1", len(transfers))
	}

	tt := transfers[0]
	if tt.Wallet != "wallet1" {
		t.Errorf("Wallet = %q", tt.Wallet)
	}
	if tt.Mint != "mintA" {
		t.Errorf("Mint = %q", tt.Mint)
	}
	if tt.Delta != 0.001 {
		t.Errorf("Delta = %f, want 0.001", tt.Delta)
	}
}

func TestComputeTokenTransfers_NewAccount(t *testing.T) {
	// No pre-balance, new token account in post
	post := []TokenBalance{
		{
			AccountIndex: 0,
			Mint:         "mintB",
			Owner:        "wallet2",
			ProgramID:    "TokenProg",
			UITokenAmount: UITokenAmount{
				Amount:         "5000",
				Decimals:       9,
				UIAmount:       ptr(0.000005),
				UIAmountString: "0.000005",
			},
		},
	}

	transfers := computeTokenTransfers(nil, post, []string{"wallet2"})
	if len(transfers) != 1 {
		t.Fatalf("transfers count = %d, want 1", len(transfers))
	}
	if transfers[0].Delta != 0.000005 {
		t.Errorf("Delta = %f", transfers[0].Delta)
	}
}

func TestComputeTokenTransfers_FullySpent(t *testing.T) {
	// Pre-balance exists, no post-balance (account closed)
	pre := []TokenBalance{
		{
			AccountIndex: 0,
			Mint:         "mintC",
			Owner:        "wallet3",
			ProgramID:    "TokenProg",
			UITokenAmount: UITokenAmount{
				Amount:         "100",
				Decimals:       6,
				UIAmount:       ptr(0.0001),
				UIAmountString: "0.0001",
			},
		},
	}

	transfers := computeTokenTransfers(pre, nil, []string{"wallet3"})
	if len(transfers) != 1 {
		t.Fatalf("transfers count = %d, want 1", len(transfers))
	}
	if transfers[0].Delta != -0.0001 {
		t.Errorf("Delta = %f, want -0.0001", transfers[0].Delta)
	}
	if transfers[0].AmountRaw != "0" {
		t.Errorf("AmountRaw = %q, want %q", transfers[0].AmountRaw, "0")
	}
}

func TestComputeTokenTransfers_NoDelta(t *testing.T) {
	// Same balance, should be skipped
	bal := TokenBalance{
		AccountIndex: 0,
		Mint:         "mintD",
		Owner:        "wallet4",
		ProgramID:    "TokenProg",
		UITokenAmount: UITokenAmount{
			Amount:         "100",
			Decimals:       6,
			UIAmount:       ptr(0.0001),
			UIAmountString: "0.0001",
		},
	}

	transfers := computeTokenTransfers([]TokenBalance{bal}, []TokenBalance{bal}, []string{"wallet4"})
	if len(transfers) != 0 {
		t.Errorf("transfers count = %d, want 0 (no change)", len(transfers))
	}
}

func TestComputeTokenTransfers_FallbackToAccountIndex(t *testing.T) {
	// Owner empty, should fall back to accountKeys[accountIndex]
	post := []TokenBalance{
		{
			AccountIndex: 1,
			Mint:         "mintE",
			Owner:        "", // empty owner
			ProgramID:    "TokenProg",
			UITokenAmount: UITokenAmount{
				Amount:         "500",
				Decimals:       6,
				UIAmount:       ptr(0.0005),
				UIAmountString: "0.0005",
			},
		},
	}

	transfers := computeTokenTransfers(nil, post, []string{"acct0", "acct1", "acct2"})
	if len(transfers) != 1 {
		t.Fatalf("transfers count = %d, want 1", len(transfers))
	}
	if transfers[0].Wallet != "acct1" {
		t.Errorf("Wallet = %q, want %q (fallback to accountKeys index)", transfers[0].Wallet, "acct1")
	}
}
