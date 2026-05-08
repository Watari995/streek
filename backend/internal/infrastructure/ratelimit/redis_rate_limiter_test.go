package ratelimit_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/Watari995/streek/backend/internal/infrastructure/ratelimit"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testKey      = "user-123"
	testRedisKey = "ratelimit:user-123"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

// expectPipelineCommands sets up the four pipeline expectations in order.
// firstSeq is the atomic counter value the limiter will use for Member (1 on first call).
func expectPipelineCommands(mock redismock.ClientMock, now time.Time, window time.Duration, firstSeq uint64, count int64) {
	windowStart := strconv.FormatInt(now.Add(-window).UnixNano()-1, 10)
	member := fmt.Sprintf("%d-%d", now.UnixNano(), firstSeq)

	mock.ExpectZRemRangeByScore(testRedisKey, "0", windowStart).SetVal(0)
	mock.ExpectZAdd(testRedisKey, redis.Z{Score: float64(now.UnixNano()), Member: member}).SetVal(1)
	mock.ExpectZCard(testRedisKey).SetVal(count)
	mock.ExpectExpire(testRedisKey, window).SetVal(true)
}

func TestRedisRateLimiter_Allow_WithinLimit_ReturnsTrue(t *testing.T) {
	t.Parallel()
	client, mock := redismock.NewClientMock()
	t.Cleanup(func() { _ = client.Close() })

	now := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	window := time.Minute
	expectPipelineCommands(mock, now, window, 1, 5)

	limiter := ratelimit.NewRedisRateLimiter(client, 10, window, fixedNow(now))
	allowed, err := limiter.Allow(context.Background(), testKey)

	require.NoError(t, err)
	assert.True(t, allowed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisRateLimiter_Allow_AtLimitBoundary_ReturnsTrue(t *testing.T) {
	t.Parallel()
	client, mock := redismock.NewClientMock()
	t.Cleanup(func() { _ = client.Close() })

	now := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	window := time.Minute
	limit := 10
	// ZADD後の件数がちょうど limit。境界はallowed.
	expectPipelineCommands(mock, now, window, 1, int64(limit))

	limiter := ratelimit.NewRedisRateLimiter(client, limit, window, fixedNow(now))
	allowed, err := limiter.Allow(context.Background(), testKey)

	require.NoError(t, err)
	assert.True(t, allowed, "count == limit should be allowed")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisRateLimiter_Allow_OverLimit_ReturnsFalse(t *testing.T) {
	t.Parallel()
	client, mock := redismock.NewClientMock()
	t.Cleanup(func() { _ = client.Close() })

	now := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	window := time.Minute
	limit := 10
	expectPipelineCommands(mock, now, window, 1, int64(limit+1))

	limiter := ratelimit.NewRedisRateLimiter(client, limit, window, fixedNow(now))
	allowed, err := limiter.Allow(context.Background(), testKey)

	require.NoError(t, err)
	assert.False(t, allowed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisRateLimiter_Allow_PipelineError_FailsOpen(t *testing.T) {
	t.Parallel()
	client, mock := redismock.NewClientMock()
	t.Cleanup(func() { _ = client.Close() })

	now := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	window := time.Minute
	windowStart := strconv.FormatInt(now.Add(-window).UnixNano()-1, 10)
	member := fmt.Sprintf("%d-1", now.UnixNano())
	wantErr := errors.New("redis down")

	mock.ExpectZRemRangeByScore(testRedisKey, "0", windowStart).SetErr(wantErr)
	mock.ExpectZAdd(testRedisKey, redis.Z{Score: float64(now.UnixNano()), Member: member}).SetErr(wantErr)
	mock.ExpectZCard(testRedisKey).SetErr(wantErr)
	mock.ExpectExpire(testRedisKey, window).SetErr(wantErr)

	limiter := ratelimit.NewRedisRateLimiter(client, 10, window, fixedNow(now))
	allowed, err := limiter.Allow(context.Background(), testKey)

	require.Error(t, err)
	assert.True(t, allowed, "fail-open: should allow request when redis errors out")
}

func TestRedisRateLimiter_Allow_IncrementsSeqAcrossCalls(t *testing.T) {
	t.Parallel()
	client, mock := redismock.NewClientMock()
	t.Cleanup(func() { _ = client.Close() })

	now := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	window := time.Minute
	// 同じ now で2回呼ぶ → seq が 1 → 2 と進むことを確認（Member の一意性担保）
	expectPipelineCommands(mock, now, window, 1, 1)
	expectPipelineCommands(mock, now, window, 2, 2)

	limiter := ratelimit.NewRedisRateLimiter(client, 10, window, fixedNow(now))
	for i := 1; i <= 2; i++ {
		allowed, err := limiter.Allow(context.Background(), testKey)
		require.NoError(t, err, "call %d", i)
		require.True(t, allowed, "call %d", i)
	}
	require.NoError(t, mock.ExpectationsWereMet())
}
