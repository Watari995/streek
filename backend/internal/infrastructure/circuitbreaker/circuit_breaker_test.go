package circuitbreaker

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExecute_Success_ResetsFailureCount(t *testing.T) {
	t.Parallel()
	cb := New("test", 3, 30*time.Second)
	assert.NoError(t, cb.Execute(func() error { return nil }))
	assert.Equal(t, 0, cb.failureCount, "should start with zero failures")

}

func TestExecute_Failure_IncrementsFailureCount(t *testing.T) {
	t.Parallel()
	cb := New("test", 3, 30*time.Second)
	// function that returns an error
	assert.Error(t, cb.Execute(func() error { return errors.New("test") }))
	assert.Equal(t, 1, cb.failureCount, "should increment failure count on failure")
}

func TestExecute_SuccessAfterFailure_Resets(t *testing.T) {
	t.Parallel()
	cb := New("test", 3, 30*time.Second)
	assert.Error(t, cb.Execute(func() error { return errors.New("test") }))
	assert.Equal(t, 1, cb.failureCount, "should increment failure count on failure")
	assert.NoError(t, cb.Execute(func() error { return nil }))
	assert.Equal(t, 0, cb.failureCount, "should reset failure count on success")
}

func TestExecute_FailureThresholdReached_OpensCircuitBreaker(t *testing.T) {
	t.Parallel()
	cb := New("test", 3, 30*time.Second)
	assert.Error(t, cb.Execute(func() error { return errors.New("test") }))
	assert.Equal(t, 1, cb.failureCount, "should increment failure count on failure")
	assert.Error(t, cb.Execute(func() error { return errors.New("test") }))
	assert.Equal(t, 2, cb.failureCount, "should increment failure count on failure")
	assert.Error(t, cb.Execute(func() error { return errors.New("test") }))
	assert.Equal(t, 3, cb.failureCount, "should increment failure count on failure")
	assert.Equal(t, StateOpen, cb.state, "should open circuit breaker on failure threshold reached")
}

func TestExecute_Open_ReturnsErrCircuitOpen(t *testing.T) {
	t.Parallel()
	fallFn := func() error { return errors.New("boom") }
	cb := New("test", 3, 30*time.Second)
	for i := 0; i < 3; i++ {
		cb.Execute(fallFn)
	}
	callCount := 0
	err := cb.Execute(func() error {
		callCount++
		return nil // 呼ばれない想定なのでなんでもいい
	})
	assert.ErrorIs(t, err, ErrCircuitOpen)
	assert.Equal(t, 0, callCount)
}

func TestExecute_HalfOpen_Success_TransitionsToClosed(t *testing.T) {
	t.Parallel()
	fakeNow := time.Now()
	cb := New("test", 3, 30*time.Second)
	cb.now = func() time.Time { return fakeNow }
	for i := 0; i < 3; i++ {
		cb.Execute(func() error { return errors.New("test") })
	}
	assert.Equal(t, StateOpen, cb.state, "should open circuit breaker on failure threshold reached")
	// reset half open when reset timeout has passed
	fakeNow = fakeNow.Add(31 * time.Second)
	assert.NoError(t, cb.Execute(func() error { return nil }))
	assert.Equal(t, StateClosed, cb.state, "should transition to closed state after success")
}

func TestExecute_HalfOpen_Failure_TransitionsToOpen(t *testing.T) {
	t.Parallel()
	fakeNow := time.Now()
	cb := New("test", 3, 30*time.Second)
	cb.now = func() time.Time { return fakeNow }
	for i := 0; i < 3; i++ {
		cb.Execute(func() error { return errors.New("test") })
	}
	fakeNow = fakeNow.Add(31 * time.Second)
	assert.Error(t, cb.Execute(func() error { return errors.New("test") }))
	assert.Equal(t, StateOpen, cb.state, "should open circuit breaker on failure threshold reached")
}
