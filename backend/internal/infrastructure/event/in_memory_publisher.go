package event

import (
	"context"
	"log"

	"github.com/Watari995/streek/backend/internal/domain/event"
)

type InMemoryPublisher struct {
	handlers      map[string][]func(ctx context.Context, event event.DomainEvent) error
	asyncHandlers map[string][]func(ctx context.Context, event event.DomainEvent) error
}

func NewInMemoryPublisher() *InMemoryPublisher {
	return &InMemoryPublisher{
		handlers:      make(map[string][]func(ctx context.Context, event event.DomainEvent) error),
		asyncHandlers: make(map[string][]func(ctx context.Context, event event.DomainEvent) error),
	}
}

func (p *InMemoryPublisher) Subscribe(eventType string, handler func(ctx context.Context, event event.DomainEvent) error) error {
	p.handlers[eventType] = append(p.handlers[eventType], handler)
	return nil
}

func (p *InMemoryPublisher) SubscribeAsync(eventType string, handler func(ctx context.Context, event event.DomainEvent) error) error {
	p.asyncHandlers[eventType] = append(p.asyncHandlers[eventType], handler)
	return nil
}

func (p *InMemoryPublisher) Publish(ctx context.Context, event event.DomainEvent) error {
	// sync handlers
	for _, h := range p.handlers[event.EventType()] {
		if err := h(ctx, event); err != nil {
			return err
		}
	}

	// async handlers
	for _, h := range p.asyncHandlers[event.EventType()] {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("async handler for %s panicked: %v", event.EventType(), r)
				}
			}()
			// use background context to avoid context cancellation
			if err := h(context.Background(), event); err != nil {
				log.Printf("async handler for %s returned error: %v", event.EventType(), err)
			}
		}()
	}
	return nil
}
