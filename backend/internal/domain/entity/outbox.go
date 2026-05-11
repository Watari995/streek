package entity

import (
	"time"

	"github.com/Watari995/streek/backend/internal/domain/valueobject"
)

const (
	outboxEventMaxRetries = 5
	outboxEventRetryCount = 0
)

type OutboxEvent struct {
	id          valueobject.OutboxEventID
	eventType   string
	payload     []byte
	status      valueobject.OutboxStatus
	retryCount  int
	maxRetries  int
	lastError   *string
	createdAt   time.Time
	processedAt *time.Time
}

func (o *OutboxEvent) ID() valueobject.OutboxEventID {
	return o.id
}

func (o *OutboxEvent) EventType() string {
	return o.eventType
}

func (o *OutboxEvent) Payload() []byte {
	return o.payload
}

func (o *OutboxEvent) Status() valueobject.OutboxStatus {
	return o.status
}

func (o *OutboxEvent) RetryCount() int {
	return o.retryCount
}

func (o *OutboxEvent) MaxRetries() int {
	return o.maxRetries
}

func (o *OutboxEvent) LastError() *string {
	return o.lastError
}

func (o *OutboxEvent) CreatedAt() time.Time {
	return o.createdAt
}

func (o *OutboxEvent) ProcessedAt() *time.Time {
	return o.processedAt
}

func NewOutboxEvent(
	id valueobject.OutboxEventID,
	eventType string,
	payload []byte,
	status valueobject.OutboxStatus,
	retryCount int,
	maxRetries int,
	lastError *string,
	createdAt time.Time,
	processedAt *time.Time,
) OutboxEvent {
	return OutboxEvent{
		id:          id,
		eventType:   eventType,
		payload:     payload,
		status:      status,
		retryCount:  retryCount,
		maxRetries:  maxRetries,
		lastError:   lastError,
		createdAt:   createdAt,
		processedAt: processedAt,
	}
}

func CreateOutboxEvent(
	eventType string,
	payload []byte,
) OutboxEvent {
	return NewOutboxEvent(
		valueobject.NewOutboxEventID(),
		eventType,
		payload,
		valueobject.NewOutboxStatusPending(),
		outboxEventRetryCount,
		outboxEventMaxRetries,
		nil,
		time.Now(),
		nil,
	)
}
