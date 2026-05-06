package circuitbreaker

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute_Success_ResetsFailureCount(t *testing.T) {
	t.Parallel()
	cb := New("test")
	assert.NoError(t, cb.Execute(func() error { return nil }))
	assert.Equal(t, 0, cb.failureCount, "should start with zero failures")

}

func TestExecute_Failure_IncrementsFailureCount(t *testing.T) {
	t.Parallel()
	cb := New("test")
	// function that returns an error
	assert.Error(t, cb.Execute(func() error { return errors.New("test") }))
	assert.Equal(t, 1, cb.failureCount, "should increment failure count on failure")
}

func TestExecute_SuccessAfterFailure_Resets(t *testing.T) {
	t.Parallel()
	cb := New("test")
	assert.Error(t, cb.Execute(func() error { return errors.New("test") }))
	assert.Equal(t, 1, cb.failureCount, "should increment failure count on failure")
	assert.NoError(t, cb.Execute(func() error { return nil }))
	assert.Equal(t, 0, cb.failureCount, "should reset failure count on success")
}
