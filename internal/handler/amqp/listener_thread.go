package amqp

import (
	"context"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
)

// [ON_THREAD_CREATED] HANDLES THREAD CREATION. RESOLVES FULL UI STATE FOR TARGETS.
func (h *MessageHandler) OnThreadCreatedV1(ctx context.Context, raw *payload.ThreadCreatedV1) ([]event.Eventer, error) {
	memberIDs := toUUIDs(raw.Members)
	var targets []uuid.UUID

	if id, err := uuid.Parse(raw.Recipient.ID); err == nil {
		if h.leader.IsLeader() || h.hub.Connected(id) {
			targets = []uuid.UUID{id}
		}
	} else {
		_, targets = h.filter("", memberIDs)
	}

	if len(targets) == 0 {
		return nil, nil
	}

	peers, err := h.enricher.Resolve(ctx, raw.DomainID, memberIDs...)
	if err != nil {
		return nil, err
	}

	base := raw.ToDomain()
	base.Members = peers

	events := make([]event.Eventer, 0, len(targets))
	for _, uid := range targets {
		thread := *base
		events = append(events, event.NewThreadEvent(&thread, uid))
	}
	return events, nil
}
