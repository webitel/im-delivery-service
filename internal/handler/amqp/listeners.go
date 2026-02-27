package amqp

import (
	"context"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
)

// [ON_MESSAGE_CREATED]
// Handles message enrichment and prepares it for distribution.
func (h *MessageHandler) OnMessageCreatedV1(ctx context.Context, userID uuid.UUID, raw *payload.MessageCreatedV1) (event.Eventer, error) {
	// [ENRICHMENT]
	// Fetch profile details for From/To entities from external services.
	from, to, err := h.enricher.ResolvePair(
		ctx,
		raw.From.ToDomain(),
		raw.To.ToDomain(),
		raw.DomainID,
	)
	if err != nil {
		h.logger.Error("PEER_ENRICHMENT_FAILED", "err", err, "msg_id", raw.MessageID)
		return nil, err // Returns err to trigger retry
	}

	// [EVENT_TRANSFORMATION]
	// Convert DTO to enriched domain event ready for WebSocket/gRPC broadcast.
	ev := event.NewMessageCreatedV1Event(raw.ToDomain(), userID, from, to)

	return ev, nil
}

// [ON_THREAD_CREATED]
func (h *MessageHandler) OnThreadCreatedV1(ctx context.Context, _ uuid.UUID, raw *payload.ThreadCreatedV1) (event.Eventer, error) {
	// Use helper to safely convert strings to UUIDs
	memberIDs := toUUIDs(raw.Members)

	// Resolve all members in one batch (cache + gRPC)
	peers, err := h.enricher.ResolveMany(ctx, memberIDs, raw.DomainID)
	if err != nil {
		h.logger.Error("THREAD_MEMBERS_ENRICHMENT_FAILED",
			"err", err,
			"thread_id", raw.ThreadID,
		)
		return nil, err
	}

	// Map raw payload to domain model and attach enriched peers
	thread := raw.ToDomain()
	thread.Members = peers

	// event.NewThreadCreatedV1Event now gets exactly what it wants: *model.Thread
	return event.NewThreadCreatedV1Event(thread), nil
}

// toUUIDs safely converts a slice of strings to UUIDs, skipping invalid ones.
func toUUIDs(ids []string) []uuid.UUID {
	res := make([]uuid.UUID, 0, len(ids))
	for _, s := range ids {
		if u, err := uuid.Parse(s); err == nil {
			res = append(res, u)
		}
	}
	return res
}

// [ON_MESSAGE_DELETED]
func (h *MessageHandler) OnMessageDeletedV1(ctx context.Context, uid uuid.UUID, raw *any) (event.Eventer, error) {
	h.logger.Debug("MOCK_DELETE_HANDLED", "user_id", uid)
	return nil, nil
}

// [ON_STATUS_CHANGED]
func (h *MessageHandler) OnStatusChangedV1(ctx context.Context, uid uuid.UUID, raw *any) (event.Eventer, error) {
	return nil, nil
}
