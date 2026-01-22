package amqp

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/service/dto"
)

func (h *MessageHandler) OnMessageCreatedV1(ctx context.Context, userID uuid.UUID, raw *dto.MessageV1) error {
	// 1. [STRICT_ENRICHMENT] Mandatory peer data resolution.
	// ResolvePeers now returns (from, to, error).
	from, to, err := h.enricher.ResolvePeers(ctx, raw.From.ToDomain(), raw.To.ToDomain(), raw.DomainID)
	if err != nil {
		// [ERROR_PROPAGATION] Returning error triggers Middleware logic.
		// It will be retried by RabbitMQ/Watermill or moved to DLX.
		return fmt.Errorf("failed to enrich participants: %w", err)
	}
	ev := event.NewMessageV1Event(raw.ToDomain(), userID, from, to)

	// 2. [LOCAL_DISPATCH] Broadcast enriched event to connected gRPC clients
	h.hub.Broadcast(ev)

	// 3. [GLOBAL_DISPATCH] Publish enriched event back to the bus
	if err := h.publisher.Publish(ctx, ev); err != nil {
		return fmt.Errorf("failed to publish enriched event: %w", err)
	}

	return nil
}
