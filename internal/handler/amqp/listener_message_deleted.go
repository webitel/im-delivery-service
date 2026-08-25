package amqp

import (
	"context"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
)

func (h *MessageHandler) OnMessageDeletedV1(ctx context.Context, raw *payload.MessageDeletedV1) ([]event.Eventer, error) {
	senderID, participantIDs := h.extractDeleteParticipants(raw)

	targets := h.computeLocalTargets(senderID, participantIDs)
	if len(targets) == 0 {
		return nil, nil
	}

	template := raw.ToDomain()

	if !template.DeletedBy.IsEnriched() {
		deleter, err := h.resolveDeleter(ctx, raw, senderID)
		if err != nil {
			return nil, err
		}

		if deleter != nil {
			template.DeletedBy = *deleter
		}
	}

	events := make([]event.Eventer, 0, len(targets))

	for _, targetID := range targets {
		isEcho := targetID == senderID
		msg := *template

		events = append(events, event.NewMessageDeletedEvent(
			&msg,
			targetID,
			event.WithEcho[*model.MessageDeleted](isEcho),
		))
	}

	return events, nil
}

// resolveDeleter covers events published without an enriched contact block:
// the membership context still comes from the event, the identity from
// im-contact-service.
func (h *MessageHandler) resolveDeleter(ctx context.Context, raw *payload.MessageDeletedV1, senderID uuid.UUID) (*model.Peer, error) {
	peers, err := h.enricher.Resolve(ctx, raw.DomainID, senderID)
	if err != nil {
		h.logger.Error("failed to enrich peer data for delete", "error", err)

		return nil, err
	}

	if len(peers) == 0 {
		return nil, nil
	}

	deleter := peers[0]
	deleter.MemberID = raw.DeletedBy.ID
	deleter.Role = int32(model.ParseRoleName(raw.DeletedBy.Role))

	return &deleter, nil
}

// extractDeleteParticipants returns the deleter plus the unique set of recipients.
func (h *MessageHandler) extractDeleteParticipants(raw *payload.MessageDeletedV1) (uuid.UUID, []uuid.UUID) {
	sID, _ := uuid.Parse(raw.DeletedBy.ContactID())
	seen := make(map[uuid.UUID]struct{})
	res := make([]uuid.UUID, 0)

	if sID != uuid.Nil {
		seen[sID] = struct{}{}
		res = append(res, sID)
	}

	for _, r := range raw.To {
		id, err := uuid.Parse(r.ContactID)
		if err == nil && id != uuid.Nil {
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				res = append(res, id)
			}
		}
	}

	return sID, res
}
