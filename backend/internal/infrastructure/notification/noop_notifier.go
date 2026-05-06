package notification

import "context"

type NoopNotifier struct{}

func NewNoopNotifier() *NoopNotifier {
	return &NoopNotifier{}
}

func (n *NoopNotifier) Notify(ctx context.Context, to string, subject string, body string) error {
	return nil
}
