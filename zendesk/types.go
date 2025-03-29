package zendesk

import "time"

type Ticket struct {
	ID           int64
	Subject      string
	Description  string
	Priority     string
	Status       string
	Requester    string
	Assignee     string
	Organization string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time

	// Comments and cursor
	Comments    []string
	AfterCursor string

	// LLM generated summary
	Summary string
}

type FetchCommentsResponse struct {
	Comments    []string
	AfterCursor string
}
