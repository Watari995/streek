//go:build integration

package ratelimit_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Watari995/streek/backend/internal/infrastructure/ratelimit"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newIntegrationClient(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	client := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		t.Skipf("redis not available at %s: %v", addr, err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// fakeClock advances time deterministically without time.Sleep.
type fakeClock struct {
	now time.Time
}

func (c *fakeClock) Now() time.Time          { return c.now }
func (c *fakeClock) Advance(d time.Duration) { c.now = c.now.Add(d) }

// uniqueUserKey avoids cross-test collision when t.Parallel() is enabled.
func uniqueUserKey(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("itest-%d", time.Now().UnixNano())
}

func cleanupKey(t *testing.T, client *redis.Client, userKey string) {
	t.Helper()
	t.Cleanup(func() {
		_ = client.Del(context.Background(), "ratelimit:"+userKey).Err()
	})
}

func TestRedisRateLimiterIntegration_AllowsUpToLimitThenBlocks(t *testing.T) {
	t.Parallel()
	client := newIntegrationClient(t)
	userKey := uniqueUserKey(t)
	cleanupKey(t, client, userKey)

	limit := 10
	clock := &fakeClock{now: time.Now()}
	limiter := ratelimit.NewRedisRateLimiter(client, limit, time.Minute, clock.Now)

	for i := 1; i <= limit; i++ {
		allowed, err := limiter.Allow(context.Background(), userKey)
		require.NoError(t, err, "iteration %d", i)
		assert.True(t, allowed, "iteration %d should be allowed", i)
	}

	allowed, err := limiter.Allow(context.Background(), userKey)
	require.NoError(t, err)
	assert.False(t, allowed, "the (limit+1)th request should be blocked")
}

func TestRedisRateLimiterIntegration_DifferentKeysAreIndependent(t *testing.T) {
	t.Parallel()
	client := newIntegrationClient(t)
	userA := uniqueUserKey(t) + "-A"
	userB := uniqueUserKey(t) + "-B"
	cleanupKey(t, client, userA)
	cleanupKey(t, client, userB)

	limit := 3
	clock := &fakeClock{now: time.Now()}
	limiter := ratelimit.NewRedisRateLimiter(client, limit, time.Minute, clock.Now)

	for i := 0; i < limit; i++ {
		allowed, err := limiter.Allow(context.Background(), userA)
		require.NoError(t, err)
		require.True(t, allowed)
	}
	allowedA, err := limiter.Allow(context.Background(), userA)
	require.NoError(t, err)
	assert.False(t, allowedA, "userA should be blocked after exceeding limit")

	allowedB, err := limiter.Allow(context.Background(), userB)
	require.NoError(t, err)
	assert.True(t, allowedB, "userB should not be affected by userA's count")
}

func TestRedisRateLimiterIntegration_ResetsAfterWindow(t *testing.T) {
	t.Parallel()
	client := newIntegrationClient(t)
	userKey := uniqueUserKey(t)
	cleanupKey(t, client, userKey)

	limit := 2
	window := time.Minute
	clock := &fakeClock{now: time.Now()}
	limiter := ratelimit.NewRedisRateLimiter(client, limit, window, clock.Now)

	for i := 0; i < limit; i++ {
		allowed, err := limiter.Allow(context.Background(), userKey)
		require.NoError(t, err)
		require.True(t, allowed)
	}

	allowed, err := limiter.Allow(context.Background(), userKey)
	require.NoError(t, err)
	assert.False(t, allowed, "should be blocked at limit+1")

	// Advance fake clock past the window — ZRemRangeByScore evicts old entries.
	clock.Advance(2 * window)

	allowed, err = limiter.Allow(context.Background(), userKey)
	require.NoError(t, err)
	assert.True(t, allowed, "should be allowed again after window elapses")
}
