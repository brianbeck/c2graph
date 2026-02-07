package handler

import (
	"net/http"

	"github.com/brianbeck/sentinel-api/internal/neo4j"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// GraphHandler handles GET /api/graph/{address} requests.
type GraphHandler struct {
	neo4jClient *neo4j.Client
}

// NewGraphHandler creates a new graph handler.
func NewGraphHandler(client *neo4j.Client) *GraphHandler {
	return &GraphHandler{neo4jClient: client}
}

// ServeHTTP handles the graph request.
func (h *GraphHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	address := chi.URLParam(r, "address")
	if !base58Regex.MatchString(address) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid Solana address"})
		return
	}

	depth := parseIntParam(r, "depth", 1)
	if depth < 1 || depth > 3 {
		depth = 1
	}

	graph, err := h.neo4jClient.GetGraph(r.Context(), address, depth)
	if err != nil {
		log.Error().Err(err).Str("address", address).Msg("Failed to fetch graph")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, graph)
}
