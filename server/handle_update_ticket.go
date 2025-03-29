package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/taonic/ticketfu/worker"
	"github.com/taonic/ticketfu/worker/ticket"
	"github.com/taonic/ticketfu/zendesk"
	"go.temporal.io/sdk/client"
)

type response struct {
	Message    string `json:"message"`
	WorkflowID string `json:"workflow_id"`
}

func (h *HTTPServer) handleUpdateTicket(w http.ResponseWriter, r *http.Request) {
	// Decode the JSON payload
	var req UpdateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	subdomain, ticketID, err := zendesk.ParseZendeskURL(req.TicketURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Received query for: subdomain: %v ticket_id: %v", subdomain, ticketID)

	// Create a unique workflow ID
	workflowID := fmt.Sprintf(ticket.TicketWorkflowIDTemplate, ticketID)

	// Create workflow input with SummarizeTicketInput struct
	input := ticket.UpsertTicketInput{
		TicketID: ticketID,
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Start or signal the workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: worker.TaskQueue,
	}

	wr, err := h.temporalClient.SignalWithStartWorkflow(
		ctx,
		workflowID,
		ticket.UpsertTicketSignal,
		input,
		workflowOptions,
		ticket.TicketWorkflow,
		nil,
	)

	if err != nil {
		http.Error(w, "Failed to start or signal workflow", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := response{
		Message:    "Ticket update workflow started or signaled",
		WorkflowID: wr.GetID(),
	}

	json.NewEncoder(w).Encode(resp)
}
