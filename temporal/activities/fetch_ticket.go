package activities

import (
	"context"
	"fmt"
	"strconv"
)

func (a *Activity) FetchTicket(ctx context.Context, id string) error {
	num, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	ticket, err := a.zClient.GetTicket(ctx, num)
	if err != nil {
		return err
	}

	fmt.Println(ticket)
	return nil
}
