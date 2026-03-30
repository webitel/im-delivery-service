package amqp

import (
	"context"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
)

// [ON_MESSAGE_CREATED] HANDLES MESSAGE FAN-OUT WITH ECHO-AWARE LOGIC.
func (h *MessageHandler) OnMessageCreatedV1(ctx context.Context, raw *payload.MessageCreatedV1) ([]event.Eventer, error) {
	senderID, participantIDs := h.extractParticipants(raw)

	targets := h.computeLocalTargets(senderID, participantIDs)
	if len(targets) == 0 {
		return nil, nil
	}

	peers, err := h.enricher.Resolve(ctx, raw.DomainID, h.getRequiredIDs(senderID, targets)...)
	if err != nil {
		return nil, err
	}

	peerMap := make(map[uuid.UUID]*model.Peer, len(peers))
	for i := range peers {
		peerMap[peers[i].ID] = &peers[i]
	}

	template := raw.ToDomain()
	template.From = *peerMap[senderID]

	events := make([]event.Eventer, 0, len(targets))
	for _, targetID := range targets {
		isEcho := targetID == senderID

		msg := *template
		msg.To = h.resolveTargetPeer(targetID, senderID, peerMap)

		events = append(events, event.NewMessageEvent(
			&msg,
			targetID,
			event.WithEcho[*model.Message](isEcho),
		))
	}

	return events, nil
}

func (h *MessageHandler) computeLocalTargets(senderID uuid.UUID, all []uuid.UUID) []uuid.UUID {
	if h.leader.IsLeader() {
		return all
	}
	res := make([]uuid.UUID, 0)
	for _, id := range all {
		if id == senderID || h.hub.Connected(id) {
			res = append(res, id)
		}
	}
	return res
}

func (h *MessageHandler) extractParticipants(raw *payload.MessageCreatedV1) (uuid.UUID, []uuid.UUID) {
	sID, _ := uuid.Parse(raw.From.ID)
	seen := make(map[uuid.UUID]struct{}, len(raw.To)+1)
	res := make([]uuid.UUID, 0, len(raw.To)+1)

	if sID != uuid.Nil {
		seen[sID], res = struct{}{}, append(res, sID)
	}
	for _, s := range raw.To {
		if id, err := uuid.Parse(s); err == nil {
			if _, ok := seen[id]; !ok {
				seen[id], res = struct{}{}, append(res, id)
			}
		}
	}
	return sID, res
}

func (h *MessageHandler) resolveTargetPeer(targetID, senderID uuid.UUID, peers map[uuid.UUID]*model.Peer) *model.Peer {
	if targetID != senderID {
		return peers[targetID]
	}
	for id, p := range peers {
		if id != senderID {
			return p
		}
	}
	return nil
}

func (h *MessageHandler) getRequiredIDs(sender uuid.UUID, targets []uuid.UUID) []uuid.UUID {
	return append([]uuid.UUID{sender}, targets...)
}
