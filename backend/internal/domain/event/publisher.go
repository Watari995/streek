package event

import "context"

type IEventPublisher interface {
	Subscribe(eventType string, handler func(ctx context.Context, event DomainEvent) error) error
	SubscribeAsync(eventType string, handler func(ctx context.Context, event DomainEvent) error) error
	Publish(ctx context.Context, event DomainEvent) error
}
