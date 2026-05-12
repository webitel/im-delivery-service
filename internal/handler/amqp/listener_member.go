package amqp

import (
	"context"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
)

// OnMemberAddedV1 handles member_added system events.
func (h *MessageHandler) OnMemberAddedV1(ctx context.Context, raw *payload.MemberEventV1) ([]event.Eventer, error) {
	return h.handleMemberEvent(raw, event.MemberAdded)
}

// OnMemberLeftV1 handles member_left system events.
func (h *MessageHandler) OnMemberLeftV1(ctx context.Context, raw *payload.MemberEventV1) ([]event.Eventer, error) {
	return h.handleMemberEvent(raw, event.MemberLeft)
}

func (h *MessageHandler) handleMemberEvent(raw *payload.MemberEventV1, kind event.EventKind) ([]event.Eventer, error) {
	m := raw.ToDomain()

	if m.ContactID.String() == "" || m.ThreadID.String() == "" {
		return nil, nil
	}

	if !h.leader.IsLeader() && !h.hub.Connected(m.ContactID) {
		return nil, nil
	}

	return []event.Eventer{event.NewMemberEvent(m, m.ContactID, kind)}, nil
}
