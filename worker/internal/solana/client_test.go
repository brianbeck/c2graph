package solana

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestClient_GetSignaturesForAddress(t *testing.T) {
	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)

		var req rpcRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Method != "getSignaturesForAddress" {
			t.Errorf("method = %q, want %q", req.Method, "getSignaturesForAddress")
		}

		resp := RPCResponse[[]SignatureInfo]{
			JSONRPC: "2.0",
			ID:      int(req.ID),
			Result: []SignatureInfo{
				{Signature: "sig1", Slot: 100},
				{Signature: "sig2", Slot: 101},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, 100)
	sigs, err := client.GetSignaturesForAddress(context.Background(), "testAddress", 10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sigs) != 2 {
		t.Fatalf("sigs count = %d, want 2", len(sigs))
	}
	if sigs[0].Signature != "sig1" {
		t.Errorf("sigs[0].Signature = %q", sigs[0].Signature)
	}
	if sigs[1].Slot != 101 {
		t.Errorf("sigs[1].Slot = %d", sigs[1].Slot)
	}
}

func TestClient_GetSignaturesForAddress_WithBefore(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		json.NewDecoder(r.Body).Decode(&req)

		// Verify "before" is in params
		if len(req.Params) < 2 {
			t.Fatal("expected 2 params")
		}
		opts, ok := req.Params[1].(map[string]interface{})
		if !ok {
			t.Fatal("params[1] should be a map")
		}
		if opts["before"] != "lastSig" {
			t.Errorf("before = %v, want %q", opts["before"], "lastSig")
		}

		resp := RPCResponse[[]SignatureInfo]{
			JSONRPC: "2.0",
			Result:  []SignatureInfo{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, 100)
	_, err := client.GetSignaturesForAddress(context.Background(), "addr", 10, "lastSig")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetTransaction(t *testing.T) {
	blockTime := int64(1700000000)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Method != "getTransaction" {
			t.Errorf("method = %q, want %q", req.Method, "getTransaction")
		}

		resp := RPCResponse[*TransactionResult]{
			JSONRPC: "2.0",
			ID:      int(req.ID),
			Result: &TransactionResult{
				Slot:      999,
				BlockTime: &blockTime,
				Transaction: &Transaction{
					Signatures: []string{"txSig123"},
					Message: Message{
						AccountKeys: []string{"walletA"},
					},
				},
				Meta: &Meta{Fee: 5000},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, 100)
	result, err := client.GetTransaction(context.Background(), "txSig123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.Slot != 999 {
		t.Errorf("Slot = %d, want 999", result.Slot)
	}
}

func TestClient_GetAccountInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Method != "getAccountInfo" {
			t.Errorf("method = %q", req.Method)
		}

		resp := RPCResponse[*AccountInfoResult]{
			JSONRPC: "2.0",
			ID:      int(req.ID),
			Result: &AccountInfoResult{
				Context: AccountInfoContext{Slot: 500},
				Value: &AccountInfoValue{
					Lamports: 1000000,
					Owner:    "11111111111111111111111111111111",
					Space:    0,
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, 100)
	info, err := client.GetAccountInfo(context.Background(), "testAddr")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil || info.Value == nil {
		t.Fatal("result should not be nil")
	}
	if info.Value.Lamports != 1000000 {
		t.Errorf("Lamports = %d", info.Value.Lamports)
	}
}

func TestClient_RPCError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"error": map[string]interface{}{
				"code":    -32600,
				"message": "Invalid request",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, 100)
	_, err := client.GetSignaturesForAddress(context.Background(), "addr", 10, "")
	if err == nil {
		t.Fatal("expected RPC error")
	}
}

func TestClient_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, 100)
	_, err := client.GetSignaturesForAddress(context.Background(), "addr", 10, "")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Slow response - won't be reached since context is cancelled
		select {}
	}))
	defer server.Close()

	client := NewClient(server.URL, 100)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.GetSignaturesForAddress(ctx, "addr", 10, "")
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestClient_GetAllSignaturesForAddress_Pagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		var sigs []SignatureInfo
		if callCount == 1 {
			// First page: return exactly 1000 (pageSize) to trigger pagination
			for i := 0; i < 1000; i++ {
				sigs = append(sigs, SignatureInfo{Signature: "sig" + string(rune('a'+i%26))})
			}
		} else {
			// Second page: return less than 1000 to stop pagination
			sigs = []SignatureInfo{{Signature: "lastSig"}}
		}

		resp := RPCResponse[[]SignatureInfo]{
			JSONRPC: "2.0",
			Result:  sigs,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, 100)
	allSigs, err := client.GetAllSignaturesForAddress(context.Background(), "addr")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 API calls (pagination), got %d", callCount)
	}
	if len(allSigs) != 1001 {
		t.Errorf("total sigs = %d, want 1001", len(allSigs))
	}
}

func TestClient_GetTokenAccountsByOwner(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Method != "getTokenAccountsByOwner" {
			t.Errorf("method = %q", req.Method)
		}

		resp := RPCResponse[*TokenAccountsResult]{
			JSONRPC: "2.0",
			ID:      int(req.ID),
			Result: &TokenAccountsResult{
				Value: []TokenAccountValue{
					{
						Pubkey: "tokenAcct1",
						Account: TokenAccountAccount{
							Lamports: 2039280,
							Owner:    "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
							Data: TokenAccountData{
								Parsed: TokenAccountParsed{
									Info: TokenAccountInfo{
										Mint:  "mintXYZ",
										Owner: "ownerAddr",
										State: "initialized",
										TokenAmount: UITokenAmount{
											Amount:         "1000000",
											Decimals:       6,
											UIAmount:       ptr(1.0),
											UIAmountString: "1",
										},
									},
								},
							},
						},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, 100)
	result, err := client.GetTokenAccountsByOwner(context.Background(), "ownerAddr")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if len(result.Value) != 1 {
		t.Fatalf("Value count = %d, want 1", len(result.Value))
	}
	if result.Value[0].Account.Data.Parsed.Info.Mint != "mintXYZ" {
		t.Errorf("Mint = %q", result.Value[0].Account.Data.Parsed.Info.Mint)
	}
}
