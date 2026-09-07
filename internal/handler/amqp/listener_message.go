package amqp

import (
	"context"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/domain/util"
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
	// The quoted message sender is enriched too, but never becomes a delivery target.
	peers, err := h.enricher.Resolve(ctx, raw.DomainID, withReplySender(participantIDs, raw)...)
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
		p.IsBot = raw.From.IsBot
	}

	// Overlay for all Recipients. IsBot is taken from the event, not the contact enricher:
	// bot-ness is a per-thread-membership property (the same contact can be a bot in one
	// thread and a plain participant in another), so the enriched per-contact IsBot may
	// disagree. The active-controller filter below depends on this being correct.
	for _, r := range raw.To {
		rid, err := uuid.Parse(r.ContactID)
		if err == nil {
			if p, ok := peerMap[rid]; ok {
				p.MemberID = r.MemberID
				p.Role = int32(r.Role)
				p.IsBot = r.IsBot
			}
		}
	}

	// A bot participant may only be woken when it is the thread's active controller
	// (bot_controller_member_id). Any other bot is suspended and must be excluded; when control
	// is released (empty controller — e.g. handed off to a human agent) no bot is woken at all.
	// This is the key routing rule: the bot-facing event published to RabbitMQ is events[0]
	// (the sender echo), whose `to` is allRecipients, and flow-manager starts a schema for
	// EVERY bot in that list — so a suspended bot (the owner while another bot or an agent
	// handles the thread) must not appear there. Humans are never filtered.
	// thread-service is the single source of truth: bot_controller_id points at the owner while
	// it should run, at the transient bot while it controls, and is NULL while an agent handles.
	controllerID := ""
	if raw.BotControllerMemberID != nil {
		controllerID = *raw.BotControllerMemberID
	}

	// System notices (member added/removed, close, ...) carry no bot_controller_member_id, so
	// the controller check would wrongly suspend every bot and strip them from the recipient
	// list. Bots must still receive system notices (a running bot reacts to a close and leaves);
	// flow-manager already refuses to START a bot from a system message, so this cannot wake a
	// suspended bot. Only regular messages are filtered to the active controller.
	isSystem := raw.System != nil || raw.Type == "system"

	suspendedBot := func(p *model.Peer) bool {
		if isSystem {
			return false
		}

		return p.IsBot && (controllerID == "" || p.MemberID != controllerID)
	}

	// Prepare the list of recipients (everyone except the sender) for Echo/System context,
	// excluding suspended bots so they are neither shown as recipients nor woken by flow.
	allRecipients := make([]model.Peer, 0)

	for _, id := range participantIDs {
		if id == senderID {
			continue
		}

		p, ok := peerMap[id]
		if !ok || suspendedBot(p) {
			continue
		}

		allRecipients = append(allRecipients, *p)
	}

	// Initialize the domain message template from the payload.
	template := raw.ToDomain()
	if sender, ok := peerMap[senderID]; ok {
		template.From = *sender
	}

	if template.ReplyTo != nil {
		if sender, ok := peerMap[template.ReplyTo.SenderID]; ok {
			template.ReplyTo.Sender = sender
		}
	}

	events := make([]event.Eventer, 0, len(targets))
	for _, targetID := range targets {
		isEcho := targetID == senderID

		if !isEcho {
			if p, ok := peerMap[targetID]; ok && suspendedBot(p) {
				continue
			}
		}

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

// withReplySender appends the quoted message sender to the enrichment list only.
func withReplySender(participantIDs []uuid.UUID, raw *payload.MessageCreatedV1) []uuid.UUID {
	if raw.ReplyTo == nil {
		return participantIDs
	}

	senderID := util.SafeParseUUID(raw.ReplyTo.SenderID)
	if senderID == uuid.Nil {
		return participantIDs
	}

	for _, id := range participantIDs {
		if id == senderID {
			return participantIDs
		}
	}

	ids := make([]uuid.UUID, len(participantIDs), len(participantIDs)+1)
	copy(ids, participantIDs)

	return append(ids, senderID)
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
