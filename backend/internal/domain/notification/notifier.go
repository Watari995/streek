package notification

import "context"

type INotifier interface {
	Notify(ctx context.Context, to string, subject string, body string) error
}
