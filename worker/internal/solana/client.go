package solana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

// endpoint holds the state for a single RPC endpoint.
type endpoint struct {
	url     string
	limiter *rate.Limiter
	healthy atomic.Bool
}

// Client is a multi-endpoint Solana JSON-RPC client with failover.
// It tries the primary endpoint first; on rate-limit (429) or error it
// falls through to the next endpoint. Each endpoint has its own rate limiter.
type Client struct {
	endpoints  []*endpoint
	httpClient *http.Client
	requestID  atomic.Int64
	mu         sync.RWMutex
	primary    int // index of current primary endpoint
}

// NewClient creates a new Solana RPC client with a single endpoint.
// Kept for backward compatibility.
func NewClient(endpointURL string, ratePerSecond int) *Client {
	return NewMultiClient([]string{endpointURL}, []int{ratePerSecond})
}

// NewMultiClient creates a client that routes requests across multiple RPC endpoints.
// The first endpoint is primary; others are fallbacks tried in order.
// ratesPerSecond must have the same length as endpoints (or will be padded with 2).
func NewMultiClient(urls []string, ratesPerSecond []int) *Client {
	eps := make([]*endpoint, 0, len(urls))
	for i, u := range urls {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}
		rps := 2
		if i < len(ratesPerSecond) && ratesPerSecond[i] > 0 {
			rps = ratesPerSecond[i]
		}
		ep := &endpoint{
			url:     u,
			limiter: rate.NewLimiter(rate.Limit(rps), rps),
		}
		ep.healthy.Store(true)
		eps = append(eps, ep)
	}

	if len(eps) == 0 {
		// Should not happen in practice; config validation catches it.
		eps = append(eps, &endpoint{
			url:     "https://api.mainnet-beta.solana.com",
			limiter: rate.NewLimiter(2, 2),
		})
		eps[0].healthy.Store(true)
	}

	log.Info().Int("endpoints", len(eps)).Strs("urls", urls).Msg("Solana RPC client initialized")

	return &Client{
		endpoints: eps,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// rpcRequest represents a JSON-RPC 2.0 request.
type rpcRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int64         `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// call routes a request through the endpoint list with per-endpoint retries.
// On 429 or connection error it immediately tries the next endpoint before
// doing any exponential backoff, maximising throughput.
func (c *Client) call(ctx context.Context, method string, params []interface{}, result interface{}) error {
	const maxRetriesPerEndpoint = 3

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

	// Build the order: start from current primary, wrap around.
	c.mu.RLock()
	startIdx := c.primary
	c.mu.RUnlock()

	var lastErr error

	for offset := 0; offset < len(c.endpoints); offset++ {
		idx := (startIdx + offset) % len(c.endpoints)
		ep := c.endpoints[idx]

		for attempt := 0; attempt <= maxRetriesPerEndpoint; attempt++ {
			// Wait for this endpoint's rate limiter
			if err := ep.limiter.Wait(ctx); err != nil {
				return fmt.Errorf("rate limiter: %w", err)
			}

			httpReq, err := http.NewRequestWithContext(ctx, "POST", ep.url, bytes.NewReader(body))
			if err != nil {
				return fmt.Errorf("create request: %w", err)
			}
			httpReq.Header.Set("Content-Type", "application/json")

			resp, err := c.httpClient.Do(httpReq)
			if err != nil {
				lastErr = fmt.Errorf("[%s] http error: %w", ep.url, err)
				if attempt < maxRetriesPerEndpoint {
					backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
					log.Warn().Err(err).Str("endpoint", ep.url).Dur("backoff", backoff).Int("attempt", attempt+1).Msg("RPC request failed, retrying")
					select {
					case <-time.After(backoff):
						continue
					case <-ctx.Done():
						return ctx.Err()
					}
				}
				// Exhausted retries on this endpoint — try next
				break
			}

			respBody, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				lastErr = fmt.Errorf("[%s] read response: %w", ep.url, err)
				break
			}

			// 429 — try next endpoint immediately, don't burn retries here
			if resp.StatusCode == http.StatusTooManyRequests {
				lastErr = fmt.Errorf("[%s] rate limited (429)", ep.url)
				log.Warn().Str("endpoint", ep.url).Msg("Rate limited (429), trying next endpoint")
				break
			}

			if resp.StatusCode != http.StatusOK {
				lastErr = fmt.Errorf("[%s] unexpected status %d: %s", ep.url, resp.StatusCode, string(respBody))
				break
			}

			// Check for JSON-RPC error
			var rpcErr struct {
				Error *RPCError `json:"error"`
			}
			if err := json.Unmarshal(respBody, &rpcErr); err == nil && rpcErr.Error != nil {
				lastErr = fmt.Errorf("[%s] RPC error %d: %s", ep.url, rpcErr.Error.Code, rpcErr.Error.Message)
				// Some RPC errors are permanent (bad params) — don't failover
				if rpcErr.Error.Code == -32600 || rpcErr.Error.Code == -32601 || rpcErr.Error.Code == -32602 {
					return lastErr
				}
				break
			}

			// Success — promote this endpoint to primary if it wasn't already
			if offset > 0 {
				c.mu.Lock()
				c.primary = idx
				c.mu.Unlock()
				log.Info().Str("endpoint", ep.url).Msg("Promoted to primary RPC endpoint")
			}

			return json.Unmarshal(respBody, result)
		}
	}

	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("all %d RPC endpoints exhausted", len(c.endpoints))
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
