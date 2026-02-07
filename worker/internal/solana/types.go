package solana

// =============================================================================
// Full-fidelity Solana JSON-RPC response types.
// These structs capture ALL fields returned by the Solana RPC API.
// =============================================================================

// --- JSON-RPC Envelope ---

// RPCResponse is the generic JSON-RPC 2.0 response wrapper.
type RPCResponse[T any] struct {
	JSONRPC string   `json:"jsonrpc"`
	ID      int      `json:"id"`
	Result  T        `json:"result"`
	Error   *RPCError `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// --- getSignaturesForAddress ---

// SignatureInfo represents one entry from getSignaturesForAddress.
type SignatureInfo struct {
	Signature          string      `json:"signature"`
	Slot               uint64      `json:"slot"`
	BlockTime          *int64      `json:"blockTime"`
	Err                interface{} `json:"err"`
	Memo               *string     `json:"memo"`
	ConfirmationStatus *string     `json:"confirmationStatus"`
}

// --- getTransaction ---

// TransactionResult is the full result from getTransaction.
type TransactionResult struct {
	Slot        uint64       `json:"slot"`
	BlockTime   *int64       `json:"blockTime"`
	Transaction *Transaction `json:"transaction"`
	Meta        *Meta        `json:"meta"`
	Version     interface{}  `json:"version"` // "legacy" (string) or version number (int)
}

// Transaction represents the transaction object (json encoding).
type Transaction struct {
	Signatures []string `json:"signatures"`
	Message    Message  `json:"message"`
}

// Message is the transaction message containing instructions and account keys.
type Message struct {
	AccountKeys         []string              `json:"accountKeys"`
	RecentBlockhash     string                `json:"recentBlockhash"`
	Header              Header                `json:"header"`
	Instructions        []Instruction         `json:"instructions"`
	AddressTableLookups []AddressTableLookup  `json:"addressTableLookups,omitempty"`
}

// Header describes the account key layout.
type Header struct {
	NumRequiredSignatures       int `json:"numRequiredSignatures"`
	NumReadonlySignedAccounts   int `json:"numReadonlySignedAccounts"`
	NumReadonlyUnsignedAccounts int `json:"numReadonlyUnsignedAccounts"`
}

// Instruction is a top-level instruction in the transaction.
type Instruction struct {
	ProgramIDIndex int    `json:"programIdIndex"`
	Accounts       []int  `json:"accounts"`
	Data           string `json:"data"`
	StackHeight    *int   `json:"stackHeight"`
}

// AddressTableLookup represents a v0 transaction address lookup table reference.
type AddressTableLookup struct {
	AccountKey      string `json:"accountKey"`
	WritableIndexes []int  `json:"writableIndexes"`
	ReadonlyIndexes []int  `json:"readonlyIndexes"`
}

// Meta contains transaction execution metadata.
type Meta struct {
	Err                  interface{}        `json:"err"`
	Fee                  uint64             `json:"fee"`
	PreBalances          []uint64           `json:"preBalances"`
	PostBalances         []uint64           `json:"postBalances"`
	PreTokenBalances     []TokenBalance     `json:"preTokenBalances"`
	PostTokenBalances    []TokenBalance     `json:"postTokenBalances"`
	InnerInstructions    []InnerInstruction `json:"innerInstructions"`
	LogMessages          []string           `json:"logMessages"`
	Rewards              []Reward           `json:"rewards"`
	LoadedAddresses      *LoadedAddresses   `json:"loadedAddresses,omitempty"`
	ReturnData           *ReturnData        `json:"returnData,omitempty"`
	ComputeUnitsConsumed *uint64            `json:"computeUnitsConsumed,omitempty"`
	Status               map[string]interface{} `json:"status"` // Deprecated but still returned
}

// TokenBalance represents a token balance entry in pre/postTokenBalances.
type TokenBalance struct {
	AccountIndex  int           `json:"accountIndex"`
	Mint          string        `json:"mint"`
	Owner         string        `json:"owner"`
	ProgramID     string        `json:"programId"`
	UITokenAmount UITokenAmount `json:"uiTokenAmount"`
}

// UITokenAmount holds human-readable token amounts.
type UITokenAmount struct {
	Amount         string   `json:"amount"`
	Decimals       int      `json:"decimals"`
	UIAmount       *float64 `json:"uiAmount"`
	UIAmountString string   `json:"uiAmountString"`
}

// InnerInstruction groups CPI instructions by the invoking top-level instruction.
type InnerInstruction struct {
	Index        int                 `json:"index"`
	Instructions []InstructionDetail `json:"instructions"`
}

// InstructionDetail is an individual inner (CPI) instruction.
type InstructionDetail struct {
	ProgramIDIndex int    `json:"programIdIndex"`
	Accounts       []int  `json:"accounts"`
	Data           string `json:"data"`
	StackHeight    *int   `json:"stackHeight"`
}

// Reward represents a reward entry in the transaction.
type Reward struct {
	Pubkey      string  `json:"pubkey"`
	Lamports    int64   `json:"lamports"`
	PostBalance uint64  `json:"postBalance"`
	RewardType  *string `json:"rewardType"`
	Commission  *uint8  `json:"commission"`
}

// LoadedAddresses contains addresses loaded via v0 address lookup tables.
type LoadedAddresses struct {
	Writable []string `json:"writable"`
	Readonly []string `json:"readonly"`
}

// ReturnData holds data returned by a program invocation.
type ReturnData struct {
	ProgramID string   `json:"programId"`
	Data      []string `json:"data"` // [base64-data, "base64"]
}

// --- getAccountInfo ---

// AccountInfoResult is the result from getAccountInfo.
type AccountInfoResult struct {
	Context AccountInfoContext `json:"context"`
	Value   *AccountInfoValue `json:"value"`
}

// AccountInfoContext provides API version and slot info.
type AccountInfoContext struct {
	APIVersion string `json:"apiVersion"`
	Slot       uint64 `json:"slot"`
}

// AccountInfoValue holds the account data fields.
type AccountInfoValue struct {
	Data       interface{} `json:"data"`       // string, [string,encoding], or parsed object
	Executable bool        `json:"executable"`
	Lamports   uint64      `json:"lamports"`
	Owner      string      `json:"owner"`
	RentEpoch  uint64      `json:"rentEpoch"`
	Space      uint64      `json:"space"`
}

// --- getTokenAccountsByOwner ---

// TokenAccountsResult is the result from getTokenAccountsByOwner.
type TokenAccountsResult struct {
	Context TokenAccountsContext `json:"context"`
	Value   []TokenAccountValue  `json:"value"`
}

// TokenAccountsContext provides API version and slot.
type TokenAccountsContext struct {
	APIVersion string `json:"apiVersion"`
	Slot       uint64 `json:"slot"`
}

// TokenAccountValue is a single token account entry.
type TokenAccountValue struct {
	Pubkey  string              `json:"pubkey"`
	Account TokenAccountAccount `json:"account"`
}

// TokenAccountAccount is the account data for a token account.
type TokenAccountAccount struct {
	Data       TokenAccountData `json:"data"`
	Executable bool             `json:"executable"`
	Lamports   uint64           `json:"lamports"`
	Owner      string           `json:"owner"`
	RentEpoch  uint64           `json:"rentEpoch"`
	Space      uint64           `json:"space"`
}

// TokenAccountData is the parsed token account data.
type TokenAccountData struct {
	Parsed  TokenAccountParsed `json:"parsed"`
	Program string             `json:"program"`
	Space   int                `json:"space"`
}

// TokenAccountParsed is the parsed info and type.
type TokenAccountParsed struct {
	Info TokenAccountInfo `json:"info"`
	Type string           `json:"type"`
}

// TokenAccountInfo contains the token balance details.
type TokenAccountInfo struct {
	IsNative    bool          `json:"isNative"`
	Mint        string        `json:"mint"`
	Owner       string        `json:"owner"`
	State       string        `json:"state"`
	TokenAmount UITokenAmount `json:"tokenAmount"`
}
