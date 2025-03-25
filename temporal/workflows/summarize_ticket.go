package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// SummarizeTicketInput is the input for the summarize ticket workflow
type SummarizeTicketInput struct {
	TicketID       string
	Summary        string
	OrganizationID string
	RequesterID    string
	RequesterEmail string
	// Sha is used to decide when to generate new summary
	Sha string
}

type SummarizeTicketOutput struct {
	Summary string
}

type summarizeTicketWorkflow struct {
	workflow.Context
	input                      SummarizeTicketInput
	signalCh                   workflow.ReceiveChannel
	updatesBeforeContinueAsNew int
	activityOptions            workflow.ActivityOptions

	// Ticket state
	compressedComments []byte
	nextComment        string
}

func newTicketSummarizer(ctx workflow.Context, input SummarizeTicketInput) *summarizeTicketWorkflow {
	return &summarizeTicketWorkflow{
		Context: workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 30 * time.Second,
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval:    time.Second,
				BackoffCoefficient: 2.0,
				MaximumInterval:    time.Minute,
				MaximumAttempts:    10,
			},
		}),
		input:    input,
		signalCh: workflow.GetSignalChannel(ctx, UpdateTicketSummarySignal),
		// We'll allow 500 updates before we continue-as-new the workflow. This is
		// required because the history will grow very large otherwise for an
		// interminable workflow fielding signal requests and executing activities.
		updatesBeforeContinueAsNew: 500,
	}
}

// Define the workflow
func SummarizeTicketWorkflow(ctx workflow.Context, input SummarizeTicketInput) error {
	s := newTicketSummarizer(ctx, input)
	return s.run()
}

func (s *summarizeTicketWorkflow) run() error {
	//a := activities.Activity{}
	//selector := workflow.NewSelector(s)
	//var wErr error

	//// Listen for cancelled
	//var cancelled bool
	//selector.AddReceive(s.Done(), func(workflow.ReceiveChannel, bool) { cancelled = true })

	//// Listen for update signals
	//var updateCount int

	//handleSignal := func(c workflow.ReceiveChannel, more bool) {
	//updateCount++
	//c.Receive(s, nil)

	//// Schedule activity to generate summary
	//input := activities.SummarizeTicketInput{
	//TicketID: s.input.TicketID,
	//Sha:      s.input.Sha,
	//}
	//selector.AddFuture(workflow.ExecuteActivity(s, a.GenerateSummaryActivity, input), func(f workflow.Future) {
	//var output activities.SummarizeTicketOutput
	//if err := f.Get(s, &output); err != nil {
	//wErr = err
	//return
	//}

	//// If sha remains the same exit early
	//if output.Sha == s.input.Sha {
	//return
	//}

	//// update summary and sha
	//s.input.Summary = output.Summary
	//s.input.Sha = output.Sha

	//// Signal org workflow through activity
	//signalOrgInput := activities.SignalOrgInput{
	//AccountID:      s.input.AccountID,
	//OrganizationID: s.input.OrganizationID,
	//TicketID:       s.input.TicketID,
	//Summary:        output.Summary,
	//}

	//// Execute signal org activity
	//future := workflow.ExecuteActivity(s, a.SignalOrgWorkflowActivity, signalOrgInput)
	//selector.AddFuture(future, func(f workflow.Future) {
	//if err := f.Get(s, nil); err != nil {
	//workflow.GetLogger(s).Error("Failed to signal org summarizer", "error", err)
	//// Note: We're not setting wErr here as we don't want to fail the main workflow
	//// if org signaling fails
	//}
	//})
	//})
	//}

	//selector.AddReceive(s.signalCh, handleSignal)
	//selector.AddReceive(s.legacySignalCh, handleSignal)

	//err := workflow.SetQueryHandler(s, internal.QueryOrgSummary, func() (SummarizeTicketOutput, error) {
	//return SummarizeTicketOutput{Summary: s.input.Summary}, nil
	//})
	//if err != nil {
	//return err
	//}

	//// Continually select until there are too many requests and no pending
	//// selects.
	////
	//// The reason we check selector.HasPending even when we've reached the request
	//// limit is to make sure no events get lost. HasPending will continually
	//// return true while an unresolved future or a buffered signal exists. If, for
	//// example, we did not check this and there was an unhandled signal buffered
	//// locally, continue-as-new would be returned without it being handled and the
	//// new workflow wouldn't get the signal either. So it'd be lost.
	//for updateCount < s.updatesBeforeContinueAsNew || selector.HasPending() {
	//selector.Select(s)
	//if cancelled {
	//return temporal.NewCanceledError()
	//}
	//if wErr != nil {
	//return wErr
	//}
	//}

	//// Continue as new since there were too many responses and the selector has
	//// nothing pending. Note, if there is request signals come in faster than they
	//// are handled or pending, there will not be a moment where the selector has
	//// nothing pending which means this will run forever.
	return workflow.NewContinueAsNewError(s, SummarizeTicketWorkflow, s.input)
}
