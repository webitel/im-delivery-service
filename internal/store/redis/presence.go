package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

const (
	// prefixPresence stores a Redis SET of active CIDs for a user.
	prefixPresence = "user:%s:presence"
	// prefixSessions stores a Redis HASH mapping CID -> DeviceID.
	prefixSessions = "user:%s:sessions"
	// prefixDevices stores a Redis HASH of serialized model.Device objects.
	prefixDevices = "user:%s:devices"
	// presenceTTL defines the automatic expiration for inactivity (missed heartbeats).
	presenceTTL = 90 * time.Second
)

type RedisPresence struct {
	rdb *redis.Client
}

func NewRedisPresence(rdb *redis.Client) *RedisPresence {
	return &RedisPresence{rdb: rdb}
}

// Online implements a dual-key registration:
// 1. Adds CID to a Set for rapid "is user online" checks.
// 2. Maps CID to DeviceID in a Hash for "which device is this" checks.
func (p *RedisPresence) Online(ctx context.Context, uid, cid uuid.UUID, deviceID string) error {
	presKey := fmt.Sprintf(prefixPresence, uid)
	sessKey := fmt.Sprintf(prefixSessions, uid)

	pipe := p.rdb.Pipeline()

	// Track connection in the user's active set
	pipe.SAdd(ctx, presKey, cid.String())
	pipe.Expire(ctx, presKey, presenceTTL)

	// Optional: link connection to a physical device ID if provided during sync
	if deviceID != "" {
		pipe.HSet(ctx, sessKey, cid.String(), deviceID)
		pipe.Expire(ctx, sessKey, presenceTTL)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Offline cleans up both the active connection set and the session mapping.
func (p *RedisPresence) Offline(ctx context.Context, uid, cid uuid.UUID) error {
	pipe := p.rdb.Pipeline()
	pipe.SRem(ctx, fmt.Sprintf(prefixPresence, uid), cid.String())
	pipe.HDel(ctx, fmt.Sprintf(prefixSessions, uid), cid.String())
	_, err := pipe.Exec(ctx)
	return err
}

// Heartbeat refreshes the TTL of the presence set to prevent premature expiration.
func (p *RedisPresence) Heartbeat(ctx context.Context, uid, cid uuid.UUID) error {
	return p.rdb.Expire(ctx, fmt.Sprintf(prefixPresence, uid), presenceTTL).Err()
}

// ActiveSessions fetches all unique connection IDs for the user.
func (p *RedisPresence) ActiveSessions(ctx context.Context, uid uuid.UUID) ([]uuid.UUID, error) {
	res, err := p.rdb.SMembers(ctx, fmt.Sprintf(prefixPresence, uid)).Result()
	if err != nil {
		return nil, err
	}

	uuids := make([]uuid.UUID, 0, len(res))
	for _, s := range res {
		if id, err := uuid.Parse(s); err == nil {
			uuids = append(uuids, id)
		}
	}
	return uuids, nil
}

// GetSessionDevice retrieves the mapped DeviceID for a specific connection CID.
func (p *RedisPresence) GetSessionDevice(ctx context.Context, uid, cid uuid.UUID) (string, error) {
	return p.rdb.HGet(ctx, fmt.Sprintf(prefixSessions, uid), cid.String()).Result()
}

// UserDevices returns all cached push-capable devices for the user.
// Returns (nil, nil) if no devices are found in the cache (Cache Miss).
func (p *RedisPresence) UserDevices(ctx context.Context, uid uuid.UUID) (*[]model.Device, error) {
	key := fmt.Sprintf(prefixDevices, uid)

	// HGetAll returns an empty map if the key does not exist (not an error).
	raw, err := p.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch devices from redis: %w", err)
	}

	// Handle Cache Miss: return nil pointer without error to trigger upstream resolution.
	if len(raw) == 0 {
		return nil, nil
	}

	devices := make([]model.Device, 0, len(raw))
	for _, val := range raw {
		var d model.Device
		if err := json.Unmarshal([]byte(val), &d); err != nil {
			continue
		}
		devices = append(devices, d)
	}

	return &devices, nil
}

// SyncDevices replaces the old device cache with a new set of data.
func (p *RedisPresence) SyncDevices(ctx context.Context, uid uuid.UUID, devices []model.Device) error {
	key := fmt.Sprintf(prefixDevices, uid)
	if len(devices) == 0 {
		return p.rdb.Del(ctx, key).Err()
	}

	fields := make(map[string]any, len(devices))
	for _, d := range devices {
		if d.ID == "" {
			continue
		}
		data, _ := json.Marshal(d)
		fields[d.ID] = data
	}

	pipe := p.rdb.Pipeline()
	pipe.Del(ctx, key) // Clear existing cache to ensure consistency
	pipe.HSet(ctx, key, fields)
	_, err := pipe.Exec(ctx)
	return err
}
