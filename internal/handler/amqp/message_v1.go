package amqp

import (
	"context"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/service/dto"
	"github.com/webitel/im-delivery-service/internal/service/mapper"
)

const (
	MessageTopicV1 = "im_message.#.message.created.v1"
	MessageQueueV1 = "im_delivery.message_v1"
)

func (h *MessageHandler) OnMessageCreatedV1(ctx context.Context, userID uuid.UUID, raw *dto.MessageV1) error {
	event := mapper.NewMessageV1Adapter(raw, userID)
	h.hub.Broadcast(event)
	return nil
}
