package cache

import (
	"context"

	domainCache "github.com/Watari995/streek/backend/internal/domain/cache"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/redis/go-redis/v9"
)

type StreakCache struct {
	client *redis.Client
}

// compile time check
var _ domainCache.IStreakCache = (*StreakCache)(nil)

func NewStreakCache(client *redis.Client) *StreakCache {
	return &StreakCache{client: client}
}

func (c *StreakCache) Get(ctx context.Context, habitID valueobject.HabitID, date valueobject.DateString) (*domainCache.StreakSnapshot, bool, error) {

}

func (c *StreakCache) Set(ctx context.Context, habitID valueobject.HabitID, date valueobject.DateString, snapshot domainCache.StreakSnapshot) error {
}

func (c *StreakCache) Invalidate(ctx context.Context, habitID valueobject.HabitID, date valueobject.DateString) error {
}
