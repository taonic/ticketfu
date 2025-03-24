package workflows

// Constants for workflow integration
const (
	TaskQueue                = "ticket-queue"
	UpdateTicketWorkflow     = "SummarizeTicketWorkflow"
	UpdateTicketStatusSignal = "update-ticket-status"
	TicketWorkflowIDTemplate = "ticket-workflow-%s" //ticket-workflow-1234 where 1234 is the ticket ID

	QueryOrgSummary           = "QueryOrgSummary"
	UpdateOrgSummarySignal    = "UpdateOrgSummary"
	UpdateTicketSummarySignal = "UpdateTicketSummary"
	OrgWorkflowIDTemplate     = "summarize-org-%s-%s" // <account-id>-<ticket-id>
)
