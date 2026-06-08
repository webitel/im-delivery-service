package amqp

import (
	"context"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
)

func (h *MessageHandler) OnInteractiveCallbackReactedV1(ctx context.Context, raw *payload.InteractiveCallbackV1) ([]event.Eventer, error) {
	domainModel, err := raw.ToDomain()
	if err != nil {
		return nil, err
	}

	peers, err := h.enricher.Resolve(ctx, 0, domainModel.ReactedBy.ID)
	if err != nil {
		return nil, err
	}

	if len(peers) != 0 {
		peer := peers[0]
		domainModel.ReactedBy = model.NewPeer(
			peer.ID,
			model.PeerUser,
			model.WithIdentity(peer.Sub, peer.Issuer, peer.Name),
			model.WithDomainID(peer.DomainID),
		)
	} else {
		h.logger.Warn(
			"can`t find internal peer information for peer",
			"contact_id",
			domainModel.ReactedBy.ID.String(),
			"in_reply_to",
			domainModel.InReplyTo,
		)
	}

	return []event.Eventer{event.NewInteractiveCallbackEvent(domainModel)}, nil
}
