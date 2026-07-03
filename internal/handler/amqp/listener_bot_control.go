package amqp

import (
	"context"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
)

// OnBotControlReleasedV1 forwards a bot control release (e.g. a user "/close") from
// im-thread-service to the im_delivery.broadcast exchange so flow_manager can stop the
// running schema. The AMQP republish itself is leader-gated by the dispatcher.
func (h *MessageHandler) OnBotControlReleasedV1(_ context.Context, raw *payload.BotControlReleasedV1) ([]event.Eventer, error) {
	b := raw.ToDomain()
	if b.ThreadID == uuid.Nil {
		return nil, nil
	}

	return []event.Eventer{event.NewBotControlReleasedEvent(b)}, nil
}
