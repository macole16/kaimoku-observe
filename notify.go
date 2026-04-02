package observe

import "context"

type NotifyProvider interface {
	Notify(ctx context.Context, userID int64, eventType, title, body string, obsID *int64) error
}
