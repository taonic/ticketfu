package workflows

import "time"

// Constants for workflow integration
const (
	UpdateTicketWorkflow     = "SummarizeTicketWorkflow"
	UpdateTicketStatusSignal = "update-ticket-status"
	TicketWorkflowIDTemplate = "ticket-workflow-%s" //ticket-workflow-1234 where 1234 is the ticket ID

	QueryOrgSummary           = "QueryOrgSummary"
	UpdateOrgSummarySignal    = "UpdateOrgSummary"
	UpdateTicketSummarySignal = "UpdateTicketSummary"
	OrgWorkflowIDTemplate     = "summarize-org-%s-%s" // <account-id>-<ticket-id>
)

type Ticket struct {
	ID             int64
	Subject        string
	Description    string
	Priority       string
	Status         string
	Submitter      string
	Assignee       string
	OrganizationID string
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// Comments and cursor
	Comments    []string
	AfterCursor string

	// LLM generated summary
	Summary string
}
