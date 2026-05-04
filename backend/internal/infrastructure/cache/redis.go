package cache

import (
	"context"
	"time"

	"github.com/Watari995/streek/backend/internal/config"
	"github.com/cockroachdb/errors"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Addr(),
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		return nil, errors.Wrap(err, "failed to ping Redis")
	}
	return rdb, nil
}
