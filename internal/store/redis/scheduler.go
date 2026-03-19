// internal/store/redis/scheduler.go
package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/webitel/im-delivery-service/internal/store"
)

const prefixScheduler = "delivery:scheduler:pending"

type RedisScheduler struct {
	rdb *redis.Client
}

func NewRedisScheduler(rdb *redis.Client) *RedisScheduler {
	return &RedisScheduler{rdb: rdb}
}

func (s *RedisScheduler) Schedule(ctx context.Context, eid, uid uuid.UUID, delay time.Duration) error {
	return s.rdb.ZAdd(ctx, prefixScheduler, redis.Z{
		Score:  float64(time.Now().Add(delay).Unix()),
		Member: fmt.Sprintf("%s:%s", eid, uid),
	}).Err()
}

func (s *RedisScheduler) PullReady(ctx context.Context) ([]store.ScheduledTask, error) {
	now := time.Now().Unix()

	// Atomic: Fetch and Delete ready tasks to prevent double-processing in a cluster.
	script := `
		local val = redis.call('ZRANGEBYSCORE', KEYS[1], '-inf', ARGV[1])
		if #val > 0 then
			redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', ARGV[1])
		end
		return val
	`
	res, err := s.rdb.Eval(ctx, script, []string{prefixScheduler}, now).Result()
	if err != nil {
		return nil, err
	}

	rawTasks := res.([]any)
	tasks := make([]store.ScheduledTask, 0, len(rawTasks))

	for _, raw := range rawTasks {
		parts := strings.Split(raw.(string), ":")
		if len(parts) == 2 {
			eid, _ := uuid.Parse(parts[0])
			uid, _ := uuid.Parse(parts[1])
			tasks = append(tasks, store.ScheduledTask{EventID: eid, UserID: uid})
		}
	}
	return tasks, nil
}
