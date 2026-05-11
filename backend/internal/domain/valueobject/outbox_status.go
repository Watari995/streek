type OutboxStatus struct {
	value string
}

const (
	outboxStatusPending = "PENDING"
	outboxStatusSent    = "SENT"
	outboxStatusFailed  = "FAILED"
)

func NewOutboxStatusPending() OutboxStatus {
	return OutboxStatus{value: outboxStatusPending}
}

func NewOutboxStatusSent() OutboxStatus {
	return OutboxStatus{value: outboxStatusSent}
}

func NewOutboxStatusFailed() OutboxStatus {
	return OutboxStatus{value: outboxStatusFailed}
}

func NewOutboxStatus(v string) (OutboxStatus, error) {
	err := validation.Validate(v,
		validation.Required,
		validation.In(outboxStatusPending, outboxStatusSent, outboxStatusFailed),
	)
	if err != nil {
		return OutboxStatus{}, err
	}
	return OutboxStatus{value: v}, nil
}

func (o OutboxStatus) String() string {
	return o.value
}

func (o OutboxStatus) IsPending() bool {
	return o.value == outboxStatusPending
}

func (o OutboxStatus) IsSent() bool {
	return o.value == outboxStatusSent
}

func (o OutboxStatus) IsFailed() bool {
	return o.value == outboxStatusFailed
}