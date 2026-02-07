package handler

import (
	"net/http"

	"github.com/brianbeck/sentinel-api/internal/neo4j"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// StatusHandler handles GET /api/status/{job_id} requests.
type StatusHandler struct {
	neo4jClient *neo4j.Client
}

// NewStatusHandler creates a new status handler.
func NewStatusHandler(client *neo4j.Client) *StatusHandler {
	return &StatusHandler{neo4jClient: client}
}

// ServeHTTP handles the status request.
func (h *StatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "job_id")
	if jobID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing job_id"})
		return
	}

	job, err := h.neo4jClient.GetScanJob(r.Context(), jobID)
	if err != nil {
		log.Error().Err(err).Str("job_id", jobID).Msg("Failed to fetch job status")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	if job == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
		return
	}

	writeJSON(w, http.StatusOK, job)
}
