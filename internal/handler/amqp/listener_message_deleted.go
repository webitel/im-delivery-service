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

	peers, err := h.enricher.Resolve(ctx, raw.DomainID, participantIDs...)
	if err != nil {
		h.logger.Error("failed to enrich peer data for delete", "error", err)

		return nil, err
	}

	peerMap := make(map[uuid.UUID]*model.Peer, len(peers))
	for i := range peers {
		peerMap[peers[i].ID] = &peers[i]
	}

	// Overlay member/role context from the raw event onto the enriched peers.
	if p, ok := peerMap[senderID]; ok {
		p.MemberID = raw.DeletedBy.MemberID
		p.Role = int32(raw.DeletedBy.Role)
	}

	for _, r := range raw.To {
		rid, err := uuid.Parse(r.ContactID)
		if err == nil {
			if p, ok := peerMap[rid]; ok {
				p.MemberID = r.MemberID
				p.Role = int32(r.Role)
			}
		}
	}

	template := raw.ToDomain()
	if sender, ok := peerMap[senderID]; ok {
		template.DeletedBy = *sender
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

// extractDeleteParticipants returns the deleter plus the unique set of recipients.
func (h *MessageHandler) extractDeleteParticipants(raw *payload.MessageDeletedV1) (uuid.UUID, []uuid.UUID) {
	sID, _ := uuid.Parse(raw.DeletedBy.ContactID)
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
