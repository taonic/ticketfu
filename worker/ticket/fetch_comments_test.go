package ticket

import (
	"errors"
	"net/http"
	"testing"

	"github.com/nukosuke/go-zendesk/zendesk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
)

func TestFetchCommentsWithActivityEnvironment(t *testing.T) {
	testSuite := testsuite.WorkflowTestSuite{}
	testEnv := testSuite.NewTestActivityEnvironment()

	testCases := []struct {
		name           string
		ticketID       string
		cursor         string
		setupMock      func(*MockZendeskClient)
		expectedOutput *FetchCommentsOutput
		expectError    bool
		errorType      string
	}{
		{
			name:     "Successful Single Page",
			ticketID: "12345",
			cursor:   "",
			setupMock: func(m *MockZendeskClient) {
				// Set up the mock to return a single page of comments
				comments := []zendesk.TicketComment{
					{PlainBody: "First comment"},
					{PlainBody: "Second comment"},
				}
				meta := zendesk.CursorPaginationMeta{
					HasMore:     false,
					AfterCursor: "cursor1",
				}
				m.On("GetTicketCommentsCBP", mock.Anything, mock.MatchedBy(func(opts *zendesk.CBPOptions) bool {
					return opts.Id == 12345 && opts.PageAfter == ""
				})).Return(comments, meta, nil).Once()
			},
			expectedOutput: &FetchCommentsOutput{
				Comments:   []string{"First comment", "Second comment"},
				NextCursor: "cursor1",
			},
		},
		{
			name:     "Multiple Pages",
			ticketID: "12345",
			cursor:   "",
			setupMock: func(m *MockZendeskClient) {
				// Set up the mock to return two pages of comments
				comments1 := []zendesk.TicketComment{
					{PlainBody: "Page 1 comment 1"},
					{PlainBody: "Page 1 comment 2"},
				}
				meta1 := zendesk.CursorPaginationMeta{
					HasMore:     true,
					AfterCursor: "cursor1",
				}
				m.On("GetTicketCommentsCBP", mock.Anything, mock.MatchedBy(func(opts *zendesk.CBPOptions) bool {
					return opts.Id == 12345 && opts.PageAfter == ""
				})).Return(comments1, meta1, nil).Once()

				comments2 := []zendesk.TicketComment{
					{PlainBody: "Page 2 comment 1"},
					{PlainBody: "Page 2 comment 2"},
				}
				meta2 := zendesk.CursorPaginationMeta{
					HasMore:     false,
					AfterCursor: "cursor2",
				}
				m.On("GetTicketCommentsCBP", mock.Anything, mock.MatchedBy(func(opts *zendesk.CBPOptions) bool {
					return opts.Id == 12345 && opts.PageAfter == "cursor1"
				})).Return(comments2, meta2, nil).Once()
			},
			expectedOutput: &FetchCommentsOutput{
				Comments: []string{
					"Page 1 comment 1", "Page 1 comment 2",
					"Page 2 comment 1", "Page 2 comment 2",
				},
				NextCursor: "cursor2",
			},
		},
		{
			name:     "Invalid Ticket ID",
			ticketID: "not-a-number",
			cursor:   "",
			setupMock: func(m *MockZendeskClient) {
				// No mock setup needed, it should fail before API call
			},
			expectError: true,
			errorType:   "InvalidArgument", // Expected error type
		},
		{
			name:     "API Error",
			ticketID: "12345",
			cursor:   "",
			setupMock: func(m *MockZendeskClient) {
				m.On("GetTicketCommentsCBP", mock.Anything, mock.Anything).
					Return([]zendesk.TicketComment{}, zendesk.CursorPaginationMeta{}, errors.New("API error")).Once()
			},
			expectError: true,
			errorType:   "ApplicationError", // Expected error type
		},
		{
			name:     "NotFound Error",
			ticketID: "99999",
			cursor:   "",
			setupMock: func(m *MockZendeskClient) {
				// Create a proper zendesk.Error using NewError
				notFoundErr := zendesk.NewError(nil, &http.Response{StatusCode: 404})
				m.On("GetTicketCommentsCBP", mock.Anything, mock.Anything).
					Return([]zendesk.TicketComment{}, zendesk.CursorPaginationMeta{}, notFoundErr).Once()
			},
			expectError: true,
			errorType:   "NotFound", // Expected error type
		},
		{
			name:     "With Initial Cursor",
			ticketID: "12345",
			cursor:   "initial-cursor",
			setupMock: func(m *MockZendeskClient) {
				comments := []zendesk.TicketComment{
					{PlainBody: "Comment with cursor"},
				}
				meta := zendesk.CursorPaginationMeta{
					HasMore:     false,
					AfterCursor: "next-cursor",
				}
				m.On("GetTicketCommentsCBP", mock.Anything, mock.MatchedBy(func(opts *zendesk.CBPOptions) bool {
					return opts.Id == 12345 && opts.PageAfter == "initial-cursor"
				})).Return(comments, meta, nil).Once()
			},
			expectedOutput: &FetchCommentsOutput{
				Comments:   []string{"Comment with cursor"},
				NextCursor: "next-cursor",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := new(MockZendeskClient)
			tc.setupMock(mockClient)

			// Create the activity instance
			activity := &Activity{
				zClient: mockClient,
			}

			// Register the activity with the test environment
			testEnv.RegisterActivity(activity.FetchComments)

			// Create the input
			input := FetchCommentsInput{
				ID:     tc.ticketID,
				Cursor: tc.cursor,
			}

			// Execute the activity
			future, err := testEnv.ExecuteActivity(activity.FetchComments, input)

			// Check for errors
			if tc.expectError {
				require.Error(t, err)
				if tc.errorType == "NotFound" {
					err := errors.Unwrap(err) // unwrap ActivityError's cause
					if appErr, ok := err.(*temporal.ApplicationError); ok {
						assert.True(t, appErr.NonRetryable())
					} else {
						assert.Fail(t, "err is not an ApplicationError", err)
					}
				} else if tc.errorType == "InvalidArgument" {
					assert.Contains(t, err.Error(), "strconv.ParseInt")
				} else {
					assert.Contains(t, err.Error(), "failed to fetch comments")
				}
			} else {
				var output FetchCommentsOutput
				err := future.Get(&output)
				require.NoError(t, err)

				assert.Equal(t, tc.expectedOutput.NextCursor, output.NextCursor)
				assert.Equal(t, tc.expectedOutput.Comments, output.Comments)
			}

			mockClient.AssertExpectations(t)
		})
	}
}
