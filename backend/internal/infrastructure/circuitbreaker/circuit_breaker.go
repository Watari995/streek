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

	mu                    sync.Mutex
	halfOpenProbeInFlight bool
}

func New(name string, failureThreshold int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{name: name, failureThreshold: failureThreshold, resetTimeout: resetTimeout, now: time.Now}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()

	if cb.state == StateOpen {
		if cb.now().Sub(cb.lastFailureAt) >= cb.resetTimeout {
			cb.state = StateHalfOpen
		} else {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
	}
	if cb.state == StateHalfOpen {
		if cb.halfOpenProbeInFlight {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
		cb.halfOpenProbeInFlight = true
	}
	cb.mu.Unlock()

	// execute function
	err := fn()
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.state == StateHalfOpen {
		cb.halfOpenProbeInFlight = false
	}
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
