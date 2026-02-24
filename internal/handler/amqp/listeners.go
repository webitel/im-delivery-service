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
	from, to, err := h.enricher.ResolvePeers(
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
	return event.NewThreadCreatedV1Event(raw.ToDomain()), nil
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
