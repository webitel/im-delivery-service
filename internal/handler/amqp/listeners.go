package amqp

import (
	"context"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
)

// [ON_MESSAGE_CREATED] Processes message creation with high-efficiency fan-out.
func (h *MessageHandler) OnMessageCreatedV1(ctx context.Context, raw *payload.MessageCreatedV1) ([]event.Eventer, error) {
	// [1. FILTER] Resolve which participants need processing on this node.
	senderID, targets := h.filter(raw.From.ID, h.extractIDs(raw))

	// [2. ENRICH] Resolve profiles in a single batch for filtered targets.
	peers, err := h.enricher.Resolve(ctx, raw.DomainID, targets...)
	if err != nil {
		h.logger.Error("failed to resolve peers", "err", err, "domain_id", raw.DomainID)
		return nil, err
	}

	// [PREPARATION] Build a lookup map for resolved peers.
	peerMap := make(map[uuid.UUID]*model.Peer)
	for i := range peers {
		// [REFERENCE_ASSIGNMENT] Take the address of the peer structure.
		peerMap[peers[i].ID] = &peers[i]
	}

	// [3. PREPARE_BASE] Map payload to domain model once.
	base := raw.ToDomain()

	// [SENDER_ATTACHMENT] Inject sender profile into the base model.
	if sender, ok := peerMap[senderID]; ok {
		// [DEREFERENCE] model.Message.From is a Value (Peer), so we dereference the pointer (*).
		base.From = *sender
	}

	// [4. FAN-OUT] Build unique events for each recipient.
	events := make([]event.Eventer, 0, len(targets))
	for _, id := range targets {
		// [SHALLOW_COPY] Create a copy of the base message structure.
		msgInstance := *base

		if recipient, ok := peerMap[id]; ok {
			// [POINTER_ASSIGNMENT] model.Message.To is a Pointer (*Peer).
			// Assign directly since recipient is already a pointer.
			msgInstance.To = recipient
		}

		events = append(events, event.NewMessageEvent(
			&msgInstance,
			id,
			event.WithPriority[*model.Message](event.PriorityHigh),
		))
	}

	return events, nil
}

// [ON_THREAD_CREATED] Handles both Direct (single recipient) and Group (fan-out) thread events.
func (h *MessageHandler) OnThreadCreatedV1(ctx context.Context, raw *payload.ThreadCreatedV1) ([]event.Eventer, error) {
	var targets []uuid.UUID

	// [1. STRATEGY_SELECTION] Determine if we target a single recipient or all members.
	// If Recipient ID is present, we treat it as a personalized "Direct" event.
	if raw.Recipient.ID != "" {
		rID, err := uuid.Parse(raw.Recipient.ID)
		if err != nil {
			h.logger.Error("failed to parse recipient id", "err", err, "id", raw.Recipient.ID)
			return nil, nil // ACK invalid data
		}

		// Only process if the recipient is connected to this node.
		if h.hub.Connected(rID) {
			targets = append(targets, rID)
		}
	} else {
		// [GROUP_LOGIC] If recipient is missing, it's a group/channel.
		// We use the standard filter to find local participants from the members list.
		_, targets = h.filter("", toUUIDs(raw.Members))
	}

	// [2. EARLY_EXIT] If no local targets found, stop processing to save resources (enrichment).
	if len(targets) == 0 {
		return nil, nil
	}

	// [3. ENRICH] Resolve all members for the UI participant list.
	allPeers, err := h.enricher.Resolve(ctx, raw.DomainID, toUUIDs(raw.Members)...)
	if err != nil {
		h.logger.Error("failed to resolve thread members", "err", err)
		return nil, err
	}

	// [4. PREPARE_BASE] Map the common thread data.
	baseThread := raw.ToDomain()
	baseThread.Members = allPeers

	// [5. DISPATCH_PREPARATION]
	events := make([]event.Eventer, 0, len(targets))
	for _, id := range targets {
		// [SHALLOW_COPY] Create a unique instance per user if needed (e.g., for logging/tracking).
		threadInstance := *baseThread

		events = append(events, event.NewThreadEvent(
			&threadInstance,
			id,
			event.WithPriority[*model.Thread](event.PriorityNormal),
		))
	}

	return events, nil
}

// [HELPERS]

// filter abstracts leadership logic: Leader processes all; Follower processes only local + sender.
func (h *MessageHandler) filter(senderIDStr string, all []uuid.UUID) (uuid.UUID, []uuid.UUID) {
	sID, _ := uuid.Parse(senderIDStr)

	// Leader Node: Responsible for global propagation via RabbitMQ.
	if h.dispatcher.IsLeader() {
		return sID, all
	}

	// Follower Node: Only handles the sender (for multi-device sync) or locally connected clients.
	filtered := make([]uuid.UUID, 0)
	for _, id := range all {
		if id == sID || h.hub.Connected(id) {
			filtered = append(filtered, id)
		}
	}
	return sID, filtered
}

// extractIDs pulls participant UUIDs from the message payload.
func (h *MessageHandler) extractIDs(raw *payload.MessageCreatedV1) []uuid.UUID {
	ids := make([]uuid.UUID, 0, 1+len(raw.To))
	if f, err := uuid.Parse(raw.From.ID); err == nil {
		ids = append(ids, f)
	}
	for _, id := range raw.To {
		if t, err := uuid.Parse(id); err == nil {
			ids = append(ids, t)
		}
	}
	return ids
}

// toUUIDs safely converts a slice of strings to UUIDs, skipping invalid formats.
func toUUIDs(src []string) []uuid.UUID {
	res := make([]uuid.UUID, 0, len(src))
	for _, s := range src {
		if u, err := uuid.Parse(s); err == nil {
			res = append(res, u)
		}
	}
	return res
}
