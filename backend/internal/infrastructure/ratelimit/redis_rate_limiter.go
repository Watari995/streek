package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/redis/go-redis/v9"
)

type RedisRateLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
	now    func() time.Time
	seq    uint64 // atomic counter
}

func NewRedisRateLimiter(client *redis.Client, limit int, window time.Duration, now func() time.Time) *RedisRateLimiter {
	return &RedisRateLimiter{
		client: client,
		limit:  limit,
		window: window,
		now:    now,
	}
}

func (r *RedisRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	now := r.now()
	windowStart := now.Add(-r.window).UnixNano()
	redisKey := "ratelimit:" + key
	member := fmt.Sprintf("%d-%d", now.UnixNano(), atomic.AddUint64(&r.seq, 1))

	// pipelineを使って4つのコマンドを1回のリクエストで実行
	pipe := r.client.Pipeline()
	// ZRemRangeByScore: 古いの削除
	pipe.ZRemRangeByScore(ctx, redisKey, "0", strconv.FormatInt(windowStart-1, 10))
	// ZAdd: 今のリクエストを追加
	pipe.ZAdd(ctx, redisKey, redis.Z{Score: float64(now.UnixNano()), Member: member})
	// ZCard: 現在のリクエスト数を取得
	countCmd := pipe.ZCard(ctx, redisKey)
	// Expire: キーの有効期限を設定
	pipe.Expire(ctx, redisKey, r.window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return true, errors.Wrap(err, "failed to execute pipeline")
	}
	return countCmd.Val() <= int64(r.limit), nil
}
