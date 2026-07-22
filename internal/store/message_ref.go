package store

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// MessageRefTracker keeps the message context of fan-out event envelopes so
// a client ACK (which references only the envelope id) can be resolved into
// a per-recipient MarkDelivered report for im-thread-service.
type MessageRefTracker interface {
	// SaveRef stores the envelope's message context with a TTL.
	SaveRef(ctx context.Context, eid uuid.UUID, ref *model.EventMessageRef, ttl time.Duration) error

	// GetRef returns the message context for the envelope, or nil when the
	// envelope is unknown (expired, or not a message event).
	GetRef(ctx context.Context, eid uuid.UUID) (*model.EventMessageRef, error)
}
