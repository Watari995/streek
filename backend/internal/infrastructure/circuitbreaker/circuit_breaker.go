package circuitbreaker

const (
	StateClosed   = "closed"
	StateOpen     = "open"
	StateHalfOpen = "half_open"
)

type CircuitBreaker struct {
	name string
}

func NewCircuitBreaker(name string) *CircuitBreaker {
	return &CircuitBreaker{name: name}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	return fn()
}
