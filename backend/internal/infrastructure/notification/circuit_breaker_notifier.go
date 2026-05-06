package notification

import (
	"context"

	domainnotification "github.com/Watari995/streek/backend/internal/domain/notification"
	"github.com/Watari995/streek/backend/internal/infrastructure/circuitbreaker"
)

type CircuitBreakerNotifier struct {
	inner domainnotification.INotifier
	cb    *circuitbreaker.CircuitBreaker
}

func NewCircuitBreakerNotifier(inner domainnotification.INotifier, cb *circuitbreaker.CircuitBreaker) *CircuitBreakerNotifier {
	return &CircuitBreakerNotifier{inner: inner, cb: cb}
}

func (n *CircuitBreakerNotifier) Notify(ctx context.Context, to string, subject string, body string) error {
	return n.cb.Execute(func() error {
		return n.inner.Notify(ctx, to, subject, body)
	})
}
