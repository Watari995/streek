package circuitbreaker

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

type CircuitBreaker struct {
	name         string
	failureCount int
}

func New(name string) *CircuitBreaker {
	return &CircuitBreaker{name: name}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	err := fn()
	if err != nil {
		cb.failureCount++
	} else {
		cb.failureCount = 0
	}
	return err
}
