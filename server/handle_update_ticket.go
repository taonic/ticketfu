package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/taonic/ticketfu/worker"
	"github.com/taonic/ticketfu/worker/ticket"
	"github.com/taonic/ticketfu/zendesk"
	"go.temporal.io/sdk/client"
	"go.temporal.io/server/common/log/tag"
)

type (
	UpdateTicketRequest struct {
		TicketURL string `json:"ticket_url"`
	}

	response struct {
		Message    string `json:"message"`
		WorkflowID string `json:"workflow_id"`
	}
)

func (h *HTTPServer) handleUpdateTicket(w http.ResponseWriter, r *http.Request) {
	var req UpdateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	_, ticketID, err := zendesk.ParseTicketURL(req.TicketURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.logger.Debug("Handling ticket update", tag.Value(ticketID))

	// Create a unique workflow ID
	workflowID := fmt.Sprintf(ticket.TicketWorkflowIDTemplate, ticketID)

	input := ticket.UpsertTicketInput{TicketID: ticketID}

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
