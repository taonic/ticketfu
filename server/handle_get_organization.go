package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/taonic/ticketfu/worker/org"
	"go.temporal.io/server/common/log/tag"
)

type GetOrganizationResponse struct {
	Summary string
}

func (h *HTTPServer) handleGetOrganization(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	h.logger.Debug("var", tag.Value(fmt.Sprintf("%+v", vars)))
	organizationId := vars["orgId"]

	// Validate required fields
	if organizationId == "" {
		h.logger.Error("Missing required field: organization_id")
		http.Error(w, "Missing organization_id in the request", http.StatusBadRequest)
		return
	}

	h.logger.Debug("Get organization", tag.Value(organizationId))

	workflowID := fmt.Sprintf(org.OrganizationWorkflowIDTemplate, organizationId)

	// Query the workflow
	val, err := h.temporalClient.QueryWorkflow(r.Context(), workflowID, "", org.QueryOrganizationSummary, "")
	if err != nil {
		h.logger.Error("Failed to query workflow", tag.Error(err))
		http.Error(w, "Failed to query workflow", http.StatusNotFound)
		return
	}

	resp := org.QueryOrganizationOutput{}
	err = val.Get(&resp)
	if err != nil {
		h.logger.Error("Failed to decode workflow response", tag.Error(err))
		http.Error(w, "Failed to decode workflow response", http.StatusInternalServerError)
		return
	}

	// Clean and parse the summary JSON
	summary := strings.TrimSpace(resp.Summary)
	summary = strings.TrimPrefix(summary, "```json")
	summary = strings.TrimSuffix(summary, "```")
	summary = strings.TrimSpace(summary)

	// Parse the summary into a generic JSON object
	var summaryJSON map[string]interface{}
	err = json.Unmarshal([]byte(summary), &summaryJSON)
	if err != nil {
		h.logger.Debug("Failed to parse summary JSON", tag.Error(err))
		http.Error(w, "Failed to parse summary JSON", http.StatusInternalServerError)
		return
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"summary": summaryJSON,
	}
	json.NewEncoder(w).Encode(response)
}
