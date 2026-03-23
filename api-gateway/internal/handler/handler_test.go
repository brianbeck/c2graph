package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// --- writeJSON tests ---

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]string{"status": "ok"}
	writeJSON(rec, http.StatusOK, data)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var result map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("body status = %q, want %q", result["status"], "ok")
	}
}

func TestWriteJSON_StatusCodes(t *testing.T) {
	codes := []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusInternalServerError, http.StatusAccepted}
	for _, code := range codes {
		rec := httptest.NewRecorder()
		writeJSON(rec, code, map[string]string{"error": "test"})
		if rec.Code != code {
			t.Errorf("writeJSON(%d) status = %d", code, rec.Code)
		}
	}
}

// --- parseIntParam tests ---

func TestParseIntParam(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		key      string
		def      int
		expected int
	}{
		{"present valid", "depth=2", "depth", 1, 2},
		{"missing", "", "depth", 1, 1},
		{"invalid", "depth=abc", "depth", 1, 1},
		{"zero", "depth=0", "depth", 1, 0},
		{"negative", "depth=-1", "depth", 1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test?"+tt.query, nil)
			result := parseIntParam(req, tt.key, tt.def)
			if result != tt.expected {
				t.Errorf("parseIntParam(%q, %q, %d) = %d, want %d", tt.query, tt.key, tt.def, result, tt.expected)
			}
		})
	}
}

// --- base58Regex tests ---

func TestBase58Regex(t *testing.T) {
	tests := []struct {
		address string
		valid   bool
	}{
		// Valid Solana addresses (32-44 chars, base58)
		{"11111111111111111111111111111111", true},   // System program (32 chars)
		{"9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM", true}, // 44 chars
		{"So11111111111111111111111111111111111111112", true},     // Wrapped SOL
		{"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA", true},   // Token program

		// Invalid
		{"", false},                   // empty
		{"short", false},              // too short
		{"0000000000000000000000000000000000000000000000", false}, // too long (46)
		{"11111111111111111111111111111111O", false},              // contains 'O' (not base58)
		{"11111111111111111111111111111111I", false},              // contains 'I' (not base58)
		{"11111111111111111111111111111111l", false},              // contains 'l' (not base58)
		{"111111111111111111111111111111110", false},              // contains '0' (not base58)
	}

	for _, tt := range tests {
		t.Run(tt.address, func(t *testing.T) {
			result := base58Regex.MatchString(tt.address)
			if result != tt.valid {
				t.Errorf("base58Regex.MatchString(%q) = %v, want %v", tt.address, result, tt.valid)
			}
		})
	}
}

// --- ScanHandler validation tests ---

func TestScanHandler_InvalidBody(t *testing.T) {
	h := &ScanHandler{} // nil deps are fine - we won't reach them

	req := httptest.NewRequest("POST", "/api/scan", strings.NewReader("not json"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestScanHandler_InvalidAddress(t *testing.T) {
	h := &ScanHandler{}

	body := `{"address": "not-valid-address!", "depth": 1}`
	req := httptest.NewRequest("POST", "/api/scan", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var result map[string]string
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["error"] != "invalid Solana address" {
		t.Errorf("error = %q, want %q", result["error"], "invalid Solana address")
	}
}

func TestScanHandler_InvalidDepth(t *testing.T) {
	h := &ScanHandler{}

	tests := []struct {
		name string
		body string
	}{
		{"depth 0", `{"address": "11111111111111111111111111111111", "depth": 0}`},
		{"depth 4", `{"address": "11111111111111111111111111111111", "depth": 4}`},
		{"depth -1", `{"address": "11111111111111111111111111111111", "depth": -1}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/scan", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}

			var result map[string]string
			json.Unmarshal(rec.Body.Bytes(), &result)
			if result["error"] != "depth must be 1, 2, or 3" {
				t.Errorf("error = %q, want %q", result["error"], "depth must be 1, 2, or 3")
			}
		})
	}
}

// --- GraphHandler validation tests ---

func TestGraphHandler_InvalidAddress(t *testing.T) {
	h := &GraphHandler{}

	// Need chi context for URL params
	r := chi.NewRouter()
	r.Get("/api/graph/{address}", h.ServeHTTP)

	req := httptest.NewRequest("GET", "/api/graph/invalid!", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// --- StatusHandler tests ---
// Note: StatusHandler requires a live Neo4j client for full testing.
// Integration tests with Neo4j are not included here.

// --- ScanRequest JSON decoding tests ---

func TestScanRequest_Decoding(t *testing.T) {
	body := `{"address": "9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM", "depth": 2}`
	var req ScanRequest
	if err := json.NewDecoder(strings.NewReader(body)).Decode(&req); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if req.Address != "9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM" {
		t.Errorf("Address = %q", req.Address)
	}
	if req.Depth != 2 {
		t.Errorf("Depth = %d, want 2", req.Depth)
	}
}

func TestScanResponse_Encoding(t *testing.T) {
	resp := ScanResponse{
		Status:  "queued",
		JobID:   "abc-123",
		Address: "9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded map[string]interface{}
	json.Unmarshal(data, &decoded)

	if decoded["status"] != "queued" {
		t.Errorf("status = %v", decoded["status"])
	}
	if decoded["job_id"] != "abc-123" {
		t.Errorf("job_id = %v", decoded["job_id"])
	}
}

func TestScanResponse_OmitEmptyJobID(t *testing.T) {
	resp := ScanResponse{
		Status:  "ready",
		Address: "9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM",
	}

	data, _ := json.Marshal(resp)
	var decoded map[string]interface{}
	json.Unmarshal(data, &decoded)

	if _, exists := decoded["job_id"]; exists {
		t.Error("job_id should be omitted when empty")
	}
}
