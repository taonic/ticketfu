package ticket

// Constants for workflow integration
const (
	TaskQueue = "ticket-queue"

	UpsertTicketSignal = "upsert-ticket-signal"

	TicketWorkflowIDTemplate = "ticket-workflow-%s" //ticket-workflow-1234 where 1234 is the ticket ID

	QueryOrgSummary        = "QueryOrgSummary"
	UpdateOrgSummarySignal = "UpdateOrgSummary"

	OrgWorkflowIDTemplate = "summarize-org-%s-%s" // <account-id>-<ticket-id>
)
