package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type DeliveryTracker interface {
	// [ACKNOWLEDGMENT] Records that a specific CID received the event (eid).
	// Also sets/refreshes the TTL for the tracking set to ensure auto-cleanup.
	Ack(ctx context.Context, eid, cid uuid.UUID, ttl time.Duration) error

	// [REPORT] Returns a list of all CIDs that confirmed receipt for a given event.
	GetAckedSessions(ctx context.Context, eid uuid.UUID) ([]uuid.UUID, error)

	// [CLEANUP] Manually removes the tracking data after the delivery flow is complete.
	Remove(ctx context.Context, eid uuid.UUID) error
}
