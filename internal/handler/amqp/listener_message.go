package amqp

import (
	"context"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
)

// OnMessageCreatedV1 handles message fan-out with contextual recipient lists.
func (h *MessageHandler) OnMessageCreatedV1(ctx context.Context, raw *payload.MessageCreatedV1) ([]event.Eventer, error) {
	senderID, participantIDs := h.extractParticipants(raw)

	// Identify targets (recipients + sender for Echo) connected to this local cluster node.
	targets := h.computeLocalTargets(senderID, participantIDs)
	if len(targets) == 0 {
		return nil, nil
	}

	// Resolve ALL participant metadata (names, types) from storage/cache.
	peers, err := h.enricher.Resolve(ctx, raw.DomainID, participantIDs...)
	if err != nil {
		h.logger.Error("failed to enrich peer data", "error", err)
		return nil, err
	}

	// Map enriched peers for quick access.
	peerMap := make(map[uuid.UUID]*model.Peer, len(peers))
	for i := range peers {
		peerMap[peers[i].ID] = &peers[i]
	}

	// [OVERLAY] Inject MemberID and Role from the RAW event into the enriched Peer objects.
	// This ensures that even if the cache doesn't have roles, the current event's context is preserved.

	// Overlay for the Sender
	if p, ok := peerMap[senderID]; ok {
		p.MemberID = raw.From.MemberID
		p.Role = int32(raw.From.Role)
	}

	// Overlay for all Recipients
	for _, r := range raw.To {
		rid, err := uuid.Parse(r.ContactID)
		if err == nil {
			if p, ok := peerMap[rid]; ok {
				p.MemberID = r.MemberID
				p.Role = int32(r.Role)
			}
		}
	}

	// Prepare the full list of recipients (everyone except the sender) for Echo/System context.
	allRecipients := make([]model.Peer, 0)
	for _, id := range participantIDs {
		if id != senderID {
			if p, ok := peerMap[id]; ok {
				allRecipients = append(allRecipients, *p)
			}
		}
	}

	// Initialize the domain message template from the payload.
	template := raw.ToDomain()
	if sender, ok := peerMap[senderID]; ok {
		template.From = *sender
	}

	events := make([]event.Eventer, 0, len(targets))
	for _, targetID := range targets {
		isEcho := targetID == senderID
		msg := *template

		if isEcho {
			// THE SENDER: Receives the full list of recipients they sent the message to.
			msg.To = allRecipients
		} else {
			// THE RECIPIENT: Receives only their own peer info in the 'To' field for privacy/clarity.
			if p, ok := peerMap[targetID]; ok {
				msg.To = []model.Peer{*p}
			} else {
				// Fallback: use allRecipients if specific peer metadata is missing.
				msg.To = allRecipients
			}
		}

		// Create the event envelope.
		// Note: The first event (i=0) in the Dispatcher will be published to RabbitMQ.
		events = append(events, event.NewMessageEvent(
			&msg,
			targetID,
			event.WithEcho[*model.Message](isEcho),
		))
	}

	return events, nil
}

// extractParticipants safely parses contact IDs and ensures unique IDs in the list.
func (h *MessageHandler) extractParticipants(raw *payload.MessageCreatedV1) (uuid.UUID, []uuid.UUID) {
	// Sender ID is parsed from ContactID as per current JSON structure.
	sID, _ := uuid.Parse(raw.From.ContactID)
	seen := make(map[uuid.UUID]struct{})
	res := make([]uuid.UUID, 0)

	if sID != uuid.Nil {
		seen[sID] = struct{}{}
		res = append(res, sID)
	}

	for _, recipient := range raw.To {
		id, err := uuid.Parse(recipient.ContactID)
		if err == nil && id != uuid.Nil {
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				res = append(res, id)
			}
		}
	}
	return sID, res
}

// computeLocalTargets determines which participants should receive the event on this node.
func (h *MessageHandler) computeLocalTargets(senderID uuid.UUID, all []uuid.UUID) []uuid.UUID {
	// Cluster leader processes all participants for global distribution/logging.
	if h.leader.IsLeader() {
		return all
	}

	// Non-leader nodes only process participants currently connected to their local WebSocket hub.
	res := make([]uuid.UUID, 0)
	for _, id := range all {
		if id == senderID || h.hub.Connected(id) {
			res = append(res, id)
		}
	}
	return res
}
