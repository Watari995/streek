package ratelimit

import "context"

type IRateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}
