package solana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

// Client is a rate-limited Solana JSON-RPC client.
type Client struct {
	endpoint   string
	httpClient *http.Client
	limiter    *rate.Limiter
	requestID  atomic.Int64
}

// NewClient creates a new Solana RPC client with rate limiting.
func NewClient(endpoint string, ratePerSecond int) *Client {
	return &Client{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		limiter: rate.NewLimiter(rate.Limit(ratePerSecond), ratePerSecond),
	}
}

// rpcRequest represents a JSON-RPC 2.0 request.
type rpcRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int64         `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// call makes a rate-limited JSON-RPC call with retry on 429.
func (c *Client) call(ctx context.Context, method string, params []interface{}, result interface{}) error {
	const maxRetries = 5

	req := rpcRequest{
		JSONRPC: "2.0",
		ID:      c.requestID.Add(1),
		Method:  method,
		Params:  params,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Wait for rate limiter
		if err := c.limiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limiter: %w", err)
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			if attempt < maxRetries {
				backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
				log.Warn().Err(err).Dur("backoff", backoff).Int("attempt", attempt+1).Msg("RPC request failed, retrying")
				select {
				case <-time.After(backoff):
					continue
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return fmt.Errorf("http request: %w", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("read response: %w", err)
		}

		// Handle rate limiting (429)
		if resp.StatusCode == http.StatusTooManyRequests {
			if attempt < maxRetries {
				backoff := time.Duration(math.Pow(2, float64(attempt+1))) * time.Second
				log.Warn().Dur("backoff", backoff).Int("attempt", attempt+1).Msg("Rate limited (429), backing off")
				select {
				case <-time.After(backoff):
					continue
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return fmt.Errorf("rate limited after %d retries", maxRetries)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
		}

		// Check for JSON-RPC error
		var rpcErr struct {
			Error *RPCError `json:"error"`
		}
		if err := json.Unmarshal(respBody, &rpcErr); err == nil && rpcErr.Error != nil {
			return fmt.Errorf("RPC error %d: %s", rpcErr.Error.Code, rpcErr.Error.Message)
		}

		return json.Unmarshal(respBody, result)
	}

	return fmt.Errorf("max retries exceeded")
}

// GetSignaturesForAddress fetches transaction signatures for a wallet address.
// Use 'before' for pagination (pass the last signature from previous page).
// Returns up to 'limit' signatures (max 1000).
func (c *Client) GetSignaturesForAddress(ctx context.Context, address string, limit int, before string) ([]SignatureInfo, error) {
	opts := map[string]interface{}{
		"limit": limit,
	}
	if before != "" {
		opts["before"] = before
	}

	params := []interface{}{address, opts}

	var resp RPCResponse[[]SignatureInfo]
	if err := c.call(ctx, "getSignaturesForAddress", params, &resp); err != nil {
		return nil, fmt.Errorf("getSignaturesForAddress(%s): %w", address, err)
	}

	return resp.Result, nil
}

// GetAllSignaturesForAddress fetches ALL transaction signatures for a wallet,
// handling pagination automatically. Returns signatures in reverse chronological order.
func (c *Client) GetAllSignaturesForAddress(ctx context.Context, address string) ([]SignatureInfo, error) {
	var allSigs []SignatureInfo
	var before string
	const pageSize = 1000

	for {
		sigs, err := c.GetSignaturesForAddress(ctx, address, pageSize, before)
		if err != nil {
			return allSigs, err
		}

		allSigs = append(allSigs, sigs...)

		if len(sigs) < pageSize {
			break // No more pages
		}

		before = sigs[len(sigs)-1].Signature
		log.Debug().Int("total", len(allSigs)).Str("address", address).Msg("Paginating signatures")
	}

	return allSigs, nil
}

// GetTransaction fetches the full transaction details for a signature.
func (c *Client) GetTransaction(ctx context.Context, signature string) (*TransactionResult, error) {
	params := []interface{}{
		signature,
		map[string]interface{}{
			"encoding":                       "json",
			"maxSupportedTransactionVersion": 0,
		},
	}

	var resp RPCResponse[*TransactionResult]
	if err := c.call(ctx, "getTransaction", params, &resp); err != nil {
		return nil, fmt.Errorf("getTransaction(%s): %w", signature, err)
	}

	return resp.Result, nil
}

// GetAccountInfo fetches account information for a public key.
func (c *Client) GetAccountInfo(ctx context.Context, address string) (*AccountInfoResult, error) {
	params := []interface{}{
		address,
		map[string]interface{}{
			"encoding": "base64",
		},
	}

	var resp RPCResponse[*AccountInfoResult]
	if err := c.call(ctx, "getAccountInfo", params, &resp); err != nil {
		return nil, fmt.Errorf("getAccountInfo(%s): %w", address, err)
	}

	return resp.Result, nil
}

// GetTokenAccountsByOwner fetches all SPL token accounts owned by the given address.
func (c *Client) GetTokenAccountsByOwner(ctx context.Context, address string) (*TokenAccountsResult, error) {
	params := []interface{}{
		address,
		map[string]interface{}{
			"programId": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
		},
		map[string]interface{}{
			"encoding": "jsonParsed",
		},
	}

	var resp RPCResponse[*TokenAccountsResult]
	if err := c.call(ctx, "getTokenAccountsByOwner", params, &resp); err != nil {
		return nil, fmt.Errorf("getTokenAccountsByOwner(%s): %w", address, err)
	}

	return resp.Result, nil
}
