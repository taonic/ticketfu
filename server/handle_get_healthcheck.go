package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status      string `json:"status"`
	TemporalOK  bool   `json:"temporal_ok"`
	TemporalMsg string `json:"temporal_msg,omitempty"`
}

func (h *HTTPServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:     "OK",
		TemporalOK: true,
	}

	// Check Temporal service health
	if h.temporalClient != nil {
		if _, err := h.temporalClient.CheckHealth(ctx, nil); err != nil {
			response.Status = "Degraded"
			response.TemporalOK = false
			response.TemporalMsg = err.Error()
		}
	} else {
		response.Status = "Degraded"
		response.TemporalOK = false
		response.TemporalMsg = "Temporal client not initialized"
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if response.Status != "OK" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(response)
}
