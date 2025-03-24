package server

// HealthResponse represents the health check response
type HealthResponse struct {
	Status      string `json:"status"`
	TemporalOK  bool   `json:"temporal_ok"`
	TemporalMsg string `json:"temporal_msg,omitempty"`
}

type UpdateTicketRequest struct {
	TicketURL      string `json:"ticket_url"`
	OrganizationID string `json:"organization_id"`
	RequesterID    string `json:"requester_id"`
	RequesterEmail string `json:"requester_email"`
}
