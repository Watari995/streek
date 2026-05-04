package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domainCache "github.com/Watari995/streek/backend/internal/domain/cache"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/cockroachdb/errors"
	"github.com/redis/go-redis/v9"
)

const streakTTL = 24 * time.Hour

type StreakCache struct {
	client *redis.Client
}

// compile time check
var _ domainCache.IStreakCache = (*StreakCache)(nil)

func NewStreakCache(client *redis.Client) *StreakCache {
	return &StreakCache{client: client}
}

func buildKey(habitID valueobject.HabitID, date valueobject.DateString) string {
	return fmt.Sprintf("streak:habit:%s:%s", habitID.String(), date.String())
}

func (c *StreakCache) Get(ctx context.Context, habitID valueobject.HabitID, date valueobject.DateString) (*domainCache.StreakSnapshot, bool, error) {

}

func (c *StreakCache) Set(ctx context.Context, habitID valueobject.HabitID, date valueobject.DateString, snapshot domainCache.StreakSnapshot) error {
	// encode snapshot to JSON
	data, err := json.Marshal(snapshot)
	if err != nil {
		return errors.Wrap(err, "failed to marshal snapshot")
	}
	// set data in Redis
	if err := c.client.Set(ctx, buildKey(habitID, date), data, streakTTL).Err(); err != nil {
		return errors.Wrap(err, "failed to set streak data in Redis")
	}
	return nil
}

func (c *StreakCache) Invalidate(ctx context.Context, habitID valueobject.HabitID, date valueobject.DateString) error {
}
