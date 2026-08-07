package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

const (
	// [PREFIX] Envelope id -> message context for ACK-to-status resolution.
	prefixEventRef = "event:msgref:%s"
)

type RedisMessageRefTracker struct {
	rdb *redis.Client
}

func NewRedisMessageRefTracker(rdb *redis.Client) *RedisMessageRefTracker {
	return &RedisMessageRefTracker{rdb: rdb}
}

// [SAVE] Persists the envelope's message context with auto-cleanup TTL.
func (t *RedisMessageRefTracker) SaveRef(ctx context.Context, eid uuid.UUID, ref *model.EventMessageRef, ttl time.Duration) error {
	raw, err := json.Marshal(ref)
	if err != nil {
		return fmt.Errorf("message_ref: marshal failed: %w", err)
	}

	return t.rdb.Set(ctx, fmt.Sprintf(prefixEventRef, eid), raw, ttl).Err()
}

// [GET] Resolves the envelope id back to its message context; nil when unknown.
func (t *RedisMessageRefTracker) GetRef(ctx context.Context, eid uuid.UUID) (*model.EventMessageRef, error) {
	raw, err := t.rdb.Get(ctx, fmt.Sprintf(prefixEventRef, eid)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("message_ref: fetch failed: %w", err)
	}

	ref := new(model.EventMessageRef)
	if err := json.Unmarshal(raw, ref); err != nil {
		return nil, fmt.Errorf("message_ref: unmarshal failed: %w", err)
	}

	return ref, nil
}
