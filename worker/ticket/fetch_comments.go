package ticket

import (
	"context"
	"fmt"
	"strconv"

	gozendesk "github.com/nukosuke/go-zendesk/zendesk"
	"github.com/taonic/ticketfu/worker/util"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
)

const (
	MaxCommentsBytes = 400 * 1024
)

type (
	FetchCommentsInput struct {
		ID     string
		Cursor string
	}

	FetchCommentsOutput struct {
		Comments   []string
		NextCursor string
	}
)

func (a *Activity) FetchComments(ctx context.Context, input FetchCommentsInput) (*FetchCommentsOutput, error) {
	logger := activity.GetLogger(ctx)

	intID, err := strconv.ParseInt(input.ID, 10, 64)
	if err != nil {
		return nil, err
	}

	var rawComments []gozendesk.TicketComment

	cpb := gozendesk.CBPOptions{
		CursorPagination: gozendesk.CursorPagination{
			PageSize:  100,
			PageAfter: input.Cursor,
		},
		CommonOptions: gozendesk.CommonOptions{
			SortOrder: "asc",
			SortBy:    "created_at",
			Id:        intID,
		},
	}

	for {
		comments, meta, err := a.zClient.GetTicketCommentsCBP(ctx, &cpb)
		if err != nil {
			if zendeskErr, ok := err.(gozendesk.Error); ok && zendeskErr.Status() == 404 {
				return nil, temporal.NewNonRetryableApplicationError("failed to find the ticket", "NotFound", err)
			}
			return nil, fmt.Errorf("failed to fetch comments: %w", err)
		}
		rawComments = append(rawComments, comments...)
		cpb.CursorPagination.PageAfter = meta.AfterCursor
		if !meta.HasMore {
			break
		}
	}

	comments := make([]string, len(rawComments))
	for i, comment := range rawComments {
		comments[i] = comment.PlainBody
	}

	// Keep the total byte size of the comments at 100KiB to reduce context window
	// todo: make it configurable
	truncatedComments, truncated := util.TruncateStringSlice(comments, MaxCommentsBytes)
	if truncated {
		logger.Debug("Truncated ticket summaries to the limit: ", MaxCommentsBytes)
	}

	response := FetchCommentsOutput{
		Comments:   truncatedComments,
		NextCursor: cpb.CursorPagination.PageAfter,
	}

	return &response, nil
}
