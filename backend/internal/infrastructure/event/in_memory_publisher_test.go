package event_test

import (
	"context"
	"errors"
	"testing"
	"time"

	domainevent "github.com/Watari995/streek/backend/internal/domain/event"
	infraevent "github.com/Watari995/streek/backend/internal/infrastructure/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubEvent is a minimal DomainEvent for testing.
type stubEvent struct {
	eventType  string
	occurredAt time.Time
}

func (e stubEvent) EventType() string     { return e.eventType }
func (e stubEvent) OccurredAt() time.Time { return e.occurredAt }

func TestInMemoryPublisher_Publish_CallsRegisteredHandler(t *testing.T) {
	t.Parallel()
	publisher := infraevent.NewInMemoryPublisher()

	called := false
	publisher.Subscribe("foo", func(ctx context.Context, e domainevent.DomainEvent) error {
		called = true
		return nil
	})

	err := publisher.Publish(context.Background(), stubEvent{eventType: "foo"})
	require.NoError(t, err)
	assert.True(t, called)
}

func TestInMemoryPublisher_Publish_PassesEventToHandler(t *testing.T) {
	t.Parallel()
	publisher := infraevent.NewInMemoryPublisher()

	want := stubEvent{eventType: "foo", occurredAt: time.Now()}
	var got domainevent.DomainEvent
	publisher.Subscribe("foo", func(ctx context.Context, e domainevent.DomainEvent) error {
		got = e
		return nil
	})

	require.NoError(t, publisher.Publish(context.Background(), want))
	assert.Equal(t, want, got)
}

func TestInMemoryPublisher_Publish_CallsAllHandlersForSameEventType(t *testing.T) {
	t.Parallel()
	publisher := infraevent.NewInMemoryPublisher()

	count := 0
	for i := 0; i < 3; i++ {
		publisher.Subscribe("foo", func(ctx context.Context, e domainevent.DomainEvent) error {
			count++
			return nil
		})
	}

	require.NoError(t, publisher.Publish(context.Background(), stubEvent{eventType: "foo"}))
	assert.Equal(t, 3, count)
}

func TestInMemoryPublisher_Publish_OnlyDispatchesMatchingEventType(t *testing.T) {
	t.Parallel()
	publisher := infraevent.NewInMemoryPublisher()

	fooCalled := false
	barCalled := false
	publisher.Subscribe("foo", func(ctx context.Context, e domainevent.DomainEvent) error {
		fooCalled = true
		return nil
	})
	publisher.Subscribe("bar", func(ctx context.Context, e domainevent.DomainEvent) error {
		barCalled = true
		return nil
	})

	require.NoError(t, publisher.Publish(context.Background(), stubEvent{eventType: "foo"}))
	assert.True(t, fooCalled)
	assert.False(t, barCalled, "handler for 'bar' should not be called when publishing 'foo'")
}

func TestInMemoryPublisher_Publish_NoHandlersIsNoOp(t *testing.T) {
	t.Parallel()
	publisher := infraevent.NewInMemoryPublisher()

	err := publisher.Publish(context.Background(), stubEvent{eventType: "unsubscribed"})
	require.NoError(t, err)
}

func TestInMemoryPublisher_Publish_StopsOnFirstHandlerError(t *testing.T) {
	t.Parallel()
	publisher := infraevent.NewInMemoryPublisher()

	wantErr := errors.New("handler failed")
	calledCount := 0
	publisher.Subscribe("foo", func(ctx context.Context, e domainevent.DomainEvent) error {
		calledCount++
		return wantErr
	})
	// This handler should NOT be called because the first one returned an error.
	publisher.Subscribe("foo", func(ctx context.Context, e domainevent.DomainEvent) error {
		calledCount++
		return nil
	})

	err := publisher.Publish(context.Background(), stubEvent{eventType: "foo"})
	require.ErrorIs(t, err, wantErr)
	assert.Equal(t, 1, calledCount, "publish should stop on first handler error")
}

func TestInMemoryPublisher_Publish_PropagatesContext(t *testing.T) {
	t.Parallel()
	publisher := infraevent.NewInMemoryPublisher()

	type ctxKey struct{}
	want := "value"

	var got any
	publisher.Subscribe("foo", func(ctx context.Context, e domainevent.DomainEvent) error {
		got = ctx.Value(ctxKey{})
		return nil
	})

	ctx := context.WithValue(context.Background(), ctxKey{}, want)
	require.NoError(t, publisher.Publish(ctx, stubEvent{eventType: "foo"}))
	assert.Equal(t, want, got)
}
