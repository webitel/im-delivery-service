package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/store"
)

type PresenceManager interface {
	// Online registers a session and optional deviceID mapping.
	Online(ctx context.Context, uid, cid uuid.UUID, deviceID string) error
	// Heartbeat refreshes the session TTL.
	Heartbeat(ctx context.Context, uid, cid uuid.UUID) error
	// Offline removes the session.
	Offline(ctx context.Context, uid, cid uuid.UUID) error
}

type PresenceService struct {
	store store.PresenceStore
}

func NewPresenceService(store store.PresenceStore) *PresenceService {
	return &PresenceService{store: store}
}

func (s *PresenceService) Online(ctx context.Context, uid, cid uuid.UUID, deviceID string) error {
	// Simply register the connection in Redis.
	return s.store.Online(ctx, uid, cid, deviceID)
}

func (s *PresenceService) Heartbeat(ctx context.Context, uid, cid uuid.UUID) error {
	// Refresh TTL without changing device mapping.
	return s.store.Online(ctx, uid, cid, "")
}

func (s *PresenceService) Offline(ctx context.Context, uid, cid uuid.UUID) error {
	return s.store.Offline(ctx, uid, cid)
}
