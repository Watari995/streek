package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type CircuitBreaker struct {
	name             string
	failureCount     int
	failureThreshold int
	state            State
	resetTimeout     time.Duration
	lastFailureAt    time.Time
	now              func() time.Time

	mu sync.Mutex
}

func New(name string, failureThreshold int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{name: name, failureThreshold: failureThreshold, resetTimeout: resetTimeout, now: time.Now}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateOpen {
		if cb.now().Sub(cb.lastFailureAt) >= cb.resetTimeout {
			cb.state = StateHalfOpen
		} else {
			return ErrCircuitOpen
		}
	}
	err := fn()
	if err != nil {
		cb.failureCount++
		if cb.failureCount >= cb.failureThreshold || cb.state == StateHalfOpen {
			cb.state = StateOpen
		}
		cb.lastFailureAt = cb.now()
	} else {
		cb.failureCount = 0
		if cb.state == StateHalfOpen {
			cb.state = StateClosed
		}
	}
	return err
}
