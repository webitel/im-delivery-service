package amqp

import (
	"context"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/service/dto"
)

// [ON_MESSAGE_CREATED]
// Handles message enrichment and prepares it for distribution.
func (h *MessageHandler) OnMessageCreatedV1(ctx context.Context, userID uuid.UUID, raw *dto.MessageV1) (event.Eventer, error) {
	// [ENRICHMENT]
	// Fetch profile details for From/To entities from external services.
	from, to, err := h.enricher.ResolvePeers(ctx, raw.From.ToDomain(), raw.To.ToDomain(), raw.DomainID)
	if err != nil {
		h.logger.Error("PEER_ENRICHMENT_FAILED", "err", err, "msg_id", raw.MessageID)
		return nil, err // Returns err to trigger retry
	}

	// [EVENT_TRANSFORMATION]
	// Convert DTO to enriched domain event ready for WebSocket/gRPC broadcast.
	return event.NewMessageV1Event(raw.ToDomain(), userID, from, to), nil
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
