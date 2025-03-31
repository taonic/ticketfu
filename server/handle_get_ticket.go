package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/taonic/ticketfu/worker/ticket"
	"go.temporal.io/server/common/log/tag"
)

type GetTicketRequest struct {
	TicketURL string `json:"ticket_url"`
}

type GetTicketResponse struct {
	Summary string
}

func (h *HTTPServer) handleGetTicket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticketID := vars["ticketId"]

	h.logger.Debug("Received get ticket", tag.Value(ticketID))

	workflowID := fmt.Sprintf(ticket.TicketWorkflowIDTemplate, ticketID)

	future, err := h.temporalClient.QueryWorkflow(r.Context(), workflowID, "", ticket.QueryTicketSummary, "")
	if err != nil {
		http.Error(w, "Failed to query workflow", http.StatusNotFound)
		log.Printf("Failed to query workflow: %v", err)
		return
	}

	output := ticket.QueryTicketOutput{}
	err = future.Get(&output)
	if err != nil {
		h.logger.Error("Failed to query workflow", tag.Error(err))
		http.Error(w, "Failed to query workflow", http.StatusNotFound)
	}

	// Send the workflow ID in the response as confirmation
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}
