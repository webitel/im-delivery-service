package amqp

import (
	"context"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
)

func (h *MessageHandler) OnInteractiveCallbackReactedV1(ctx context.Context, raw *payload.InteractiveCallbackV1) ([]event.Eventer, error) {
	if !h.leader.IsLeader() {
		return []event.Eventer{}, nil
	}

	domainModel, err := raw.ToDomain()
	if err != nil {
		return nil, err
	}

	peers, err := h.enricher.Resolve(ctx, int32(domainModel.DomainID), domainModel.ReactedBy.ID, domainModel.Receiver.ID)
	if err != nil {
		return nil, err
	}

	for _, peer := range peers {
		if peer.ID == domainModel.ReactedBy.ID {
			domainModel.ReactedBy = peer
		}

		if peer.ID == domainModel.Receiver.ID {
			domainModel.Receiver = peer
		}
	}

	return []event.Eventer{event.NewInteractiveCallbackEvent(domainModel)}, nil
}
