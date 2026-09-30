package amqp

import (
	"context"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
)

// OnMessageStatusV1 fans a per-recipient delivery status change out to all
// thread participants connected to this node, so UI/SDK clients update the
// status marks (checkmarks) in real time.
func (h *MessageHandler) OnMessageStatusV1(ctx context.Context, raw *payload.MessageStatusV1) ([]event.Eventer, error) {
	participants := raw.ParticipantIDs()
	if len(participants) == 0 {
		return nil, nil
	}

	// Status events have no "sender": every participant is a plain target.
	targets := h.computeLocalTargets(uuid.Nil, participants)
	if len(targets) == 0 {
		return nil, nil
	}

	update := raw.ToDomain()
	update.Member = h.statusMember(ctx, raw, update.MemberID)

	events := make([]event.Eventer, 0, len(targets))
	for _, targetID := range targets {
		events = append(events, event.NewMessageStatusEvent(update, targetID, int64(raw.DomainID)))
	}

	return events, nil
}

// statusMember resolves the recipient into a full member, the way a message sender is: the
// membership comes from the event, the identity from im-contact-service.
func (h *MessageHandler) statusMember(ctx context.Context, raw *payload.MessageStatusV1, contactID uuid.UUID) *model.Peer {
	member := model.Peer{ID: contactID}

	if peers, err := h.enricher.Resolve(ctx, raw.DomainID, contactID); err != nil {
		h.logger.Warn("failed to enrich status member", "error", err)
	} else if len(peers) > 0 {
		member = peers[0]
	}

	if raw.Member != nil {
		member.MemberID = raw.Member.MemberID
		member.Role = int32(raw.Member.Role)
		member.IsBot = raw.Member.IsBot
	}

	return &member
}
