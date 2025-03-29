package ticket

import (
	"context"
	"fmt"
	"strconv"

	gozendesk "github.com/nukosuke/go-zendesk/zendesk"
	"github.com/taonic/ticketfu/zendesk"
	"go.temporal.io/sdk/temporal"
)

func (a *Activity) FetchComments(ctx context.Context, id string, cursor string) (*zendesk.FetchCommentsResponse, error) {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}

	var allComments []gozendesk.TicketComment

	cpb := gozendesk.CBPOptions{
		CursorPagination: gozendesk.CursorPagination{
			PageSize:  100,
			PageAfter: cursor,
		},
		CommonOptions: gozendesk.CommonOptions{
			SortOrder: "asc",
			SortBy:    "created_at",
			Id:        intID,
		},
	}

	fmt.Println(cpb)

	for {
		comments, meta, err := a.zClient.GetTicketCommentsCBP(ctx, &cpb)
		if err != nil {
			if zendeskErr, ok := err.(gozendesk.Error); ok && zendeskErr.Status() == 404 {
				return nil, temporal.NewNonRetryableApplicationError("failed to find the ticket", "NotFound", err)
			}
			return nil, fmt.Errorf("failed to fetch comments: %w", err)
		}
		allComments = append(allComments, comments...)
		cpb.CursorPagination.PageAfter = meta.AfterCursor
		if !meta.HasMore {
			break
		}
	}

	comments := make([]string, len(allComments))
	for i, comment := range allComments {
		comments[i] = comment.PlainBody
	}

	response := zendesk.FetchCommentsResponse{
		Comments:    comments,
		AfterCursor: cpb.CursorPagination.PageAfter,
	}

	return &response, nil
}
