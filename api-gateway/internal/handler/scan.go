package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"

	"github.com/brianbeck/sentinel-api/internal/neo4j"
	"github.com/brianbeck/sentinel-api/internal/queue"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

var base58Regex = regexp.MustCompile(`^[1-9A-HJ-NP-Za-km-z]{32,44}$`)

// ScanHandler handles POST /api/scan requests.
type ScanHandler struct {
	neo4jClient   *neo4j.Client
	publisher     *queue.Publisher
	maxWalletsCap int
}

// NewScanHandler creates a new scan handler.
func NewScanHandler(client *neo4j.Client, pub *queue.Publisher, maxWalletsCap int) *ScanHandler {
	return &ScanHandler{
		neo4jClient:   client,
		publisher:     pub,
		maxWalletsCap: maxWalletsCap,
	}
}

// ScanRequest is the expected JSON body for POST /api/scan.
type ScanRequest struct {
	Address string `json:"address"`
	Depth   int    `json:"depth"`
}

// ScanResponse is returned from POST /api/scan.
type ScanResponse struct {
	Status  string `json:"status"`
	JobID   string `json:"job_id,omitempty"`
	Address string `json:"address"`
}

// ServeHTTP handles the scan request.
func (h *ScanHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Validate address
	if !base58Regex.MatchString(req.Address) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid Solana address"})
		return
	}

	// Validate depth
	if req.Depth < 1 || req.Depth > 3 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "depth must be 1, 2, or 3"})
		return
	}

	// Check if data is fresh (Layer 1 dedup)
	fresh, err := h.neo4jClient.IsWalletFresh(r.Context(), req.Address, req.Depth)
	if err != nil {
		log.Error().Err(err).Str("address", req.Address).Msg("Freshness check failed")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	if fresh {
		writeJSON(w, http.StatusOK, ScanResponse{
			Status:  "ready",
			Address: req.Address,
		})
		return
	}

	// Create a scan job and publish to queue
	jobID := uuid.New().String()

	if err := h.neo4jClient.CreateScanJob(r.Context(), jobID, req.Address, req.Depth, h.maxWalletsCap); err != nil {
		log.Error().Err(err).Msg("Failed to create scan job")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	msg := &queue.ScanMessage{
		JobID:        jobID,
		Address:      req.Address,
		Depth:        req.Depth,
		CurrentDepth: req.Depth,
		RootJobID:    jobID,
	}

	if err := h.publisher.Publish(r.Context(), msg); err != nil {
		log.Error().Err(err).Str("job_id", jobID).Msg("Failed to publish scan request")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to queue scan"})
		return
	}

	log.Info().Str("job_id", jobID).Str("address", req.Address).Int("depth", req.Depth).Msg("Scan queued")

	writeJSON(w, http.StatusAccepted, ScanResponse{
		Status:  "queued",
		JobID:   jobID,
		Address: req.Address,
	})
}

// used by scan.go - intentionally here since it's the first handler
// but shared across handlers
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func parseIntParam(r *http.Request, key string, defaultVal int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return i
}
