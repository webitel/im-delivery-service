// internal/store/redis/scheduler.go
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/webitel/im-delivery-service/internal/domain/event"
)

const (
	prefixScheduler = "delivery:scheduler:pending"
	fmtEventKey     = "delivery:event:%s"
)

type RedisScheduler struct {
	rdb *redis.Client
}

func NewRedisScheduler(rdb *redis.Client) *RedisScheduler {
	return &RedisScheduler{rdb: rdb}
}

// [SCHEDULE] Atomically stores event data and sets execution time.
func (s *RedisScheduler) Schedule(ctx context.Context, ev event.Eventer, delay time.Duration) error {
	eid := ev.GetID()
	uid := ev.GetUserID().String()

	data, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("marshal_failed: %w", err)
	}

	eventKey := fmt.Sprintf(fmtEventKey, eid)
	// [TTL] Safety margin for orphan data (24h).
	// Lua script will delete it much sooner (immediately after processing).
	eventTTL := delay + (24 * time.Hour)

	pipe := s.rdb.Pipeline()
	pipe.Set(ctx, eventKey, data, eventTTL)
	pipe.ZAdd(ctx, prefixScheduler, redis.Z{
		Score:  float64(time.Now().Add(delay).Unix()),
		Member: fmt.Sprintf("%s:%s", eid, uid),
	})

	_, err = pipe.Exec(ctx)
	return err
}

// [PULL_READY] The Atomic Unified Fetcher.
func (s *RedisScheduler) PullReady(ctx context.Context) ([]event.Eventer, error) {
	now := time.Now().Unix()

	// [LUA] This script guarantees that data is pulled and deleted in one step.
	script := `
		local members = redis.call('ZRANGEBYSCORE', KEYS[1], '-inf', ARGV[1])
		if #members == 0 then return {} end
		
		local payloads = {}
		for i, member in ipairs(members) do
			-- Extract eid from "eid:uid" string
			local eid = string.match(member, "([^:]+)")
			local key = "delivery:event:" .. eid
			
			payloads[i] = redis.call('GET', key)
			redis.call('DEL', key) -- Immediate cleanup
		end
		
		redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', ARGV[1])
		return payloads
	`

	res, err := s.rdb.Eval(ctx, script, []string{prefixScheduler}, now).Result()
	if err != nil {
		return nil, err
	}

	rawPayloads := res.([]any)
	events := make([]event.Eventer, 0, len(rawPayloads))

	for _, raw := range rawPayloads {
		if raw == nil {
			continue
		}
		data := []byte(raw.(string))

		// [POLYMORPHIC_RESTORE] Detect kind and create concrete struct.
		var meta struct {
			Kind event.EventKind `json:"kind"`
		}
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}

		ev := event.NewEnvelopeForKind(meta.Kind)
		if ev == nil {
			continue
		}

		if err := json.Unmarshal(data, ev); err != nil {
			continue
		}
		events = append(events, ev)
	}

	return events, nil
}
