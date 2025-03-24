package server

// HealthResponse represents the health check response
type HealthResponse struct {
	Status      string `json:"status"`
	TemporalOK  bool   `json:"temporal_ok"`
	TemporalMsg string `json:"temporal_msg,omitempty"`
}
