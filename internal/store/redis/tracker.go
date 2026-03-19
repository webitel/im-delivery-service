package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// [PREFIX] Unique Set per Event ID to track global ACKs.
	prefixAckSet = "event:acks:%s"
)

type RedisTracker struct {
	rdb *redis.Client
}

func NewRedisTracker(rdb *redis.Client) *RedisTracker {
	return &RedisTracker{rdb: rdb}
}

// [ACK] Records a unique connection ID that confirmed message receipt.
func (t *RedisTracker) Ack(ctx context.Context, eid, cid uuid.UUID, ttl time.Duration) error {
	key := fmt.Sprintf(prefixAckSet, eid)

	// [PIPELINE] SADD and EXPIRE combined to ensure cleanup even if process fails.
	pipe := t.rdb.Pipeline()
	pipe.SAdd(ctx, key, cid.String())
	pipe.Expire(ctx, key, ttl)

	_, err := pipe.Exec(ctx)
	return err
}

// [REPORT] Collects all session IDs that acknowledged the event.
func (t *RedisTracker) GetAckedSessions(ctx context.Context, eid uuid.UUID) ([]uuid.UUID, error) {
	key := fmt.Sprintf(prefixAckSet, eid)

	// [FETCH] SMEMBERS returns the full set of unique CIDs.
	res, err := t.rdb.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("tracker: failed to fetch ACKs: %w", err)
	}

	// [PARSE] Convert Redis strings back to internal UUID type.
	acked := make([]uuid.UUID, 0, len(res))
	for _, s := range res {
		if id, err := uuid.Parse(s); err == nil {
			acked = append(acked, id)
		}
	}
	return acked, nil
}

// [CLEANUP] Explicitly removes the tracking set once the push logic is completed.
func (t *RedisTracker) Remove(ctx context.Context, eid uuid.UUID) error {
	key := fmt.Sprintf(prefixAckSet, eid)
	return t.rdb.Del(ctx, key).Err()
}
