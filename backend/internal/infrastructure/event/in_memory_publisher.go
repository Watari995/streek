package event

import (
	"context"

	"github.com/Watari995/streek/backend/internal/domain/event"
)

type InMemoryPublisher struct {
	handlers map[string][]func(ctx context.Context, event event.DomainEvent) error
}

func NewInMemoryPublisher() *InMemoryPublisher {
	return &InMemoryPublisher{
		handlers: make(map[string][]func(ctx context.Context, event event.DomainEvent) error),
	}
}

func (p *InMemoryPublisher) Subscribe(eventType string, handler func(ctx context.Context, event event.DomainEvent) error) error {
	p.handlers[eventType] = append(p.handlers[eventType], handler)
	return nil
}

func (p *InMemoryPublisher) Publish(ctx context.Context, event event.DomainEvent) error {
	for _, h := range p.handlers[event.EventType()] {
		if err := h(ctx, event); err != nil {
			return err
		}
	}
	return nil
}
