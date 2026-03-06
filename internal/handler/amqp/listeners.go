package amqp

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
)

// [ON_MESSAGE_CREATED] Handles message fan-out with smart enrichment for local targets.
func (h *MessageHandler) OnMessageCreatedV1(ctx context.Context, raw *payload.MessageCreatedV1) ([]event.Eventer, error) {
	sID, participants := extractParticipants(raw)

	// [FILTER] Identify who on THIS node should receive the event
	_, targets := h.filter(raw.From.ID, participants)
	if len(targets) == 0 {
		return nil, nil
	}

	// [ENRICHMENT_LOGIC] Determine the minimal set of peers to resolve
	toResolve := targets
	isSenderLocal := slices.Contains(targets, sID)

	if isSenderLocal {
		// [SENDER_ECHO] If sender is local, we need everyone to populate the "To" field correctly
		toResolve = participants
	} else {
		// [RECIPIENT_ONLY] If only recipients are local, we need them + the sender for the "From" field
		if sID != uuid.Nil && !slices.Contains(targets, sID) {
			toResolve = append(slices.Clone(targets), sID)
		}
	}

	// [RESOLVE] Fetch metadata for the required participants
	peers, err := h.enricher.Resolve(ctx, raw.DomainID, toResolve...)
	if err != nil {
		return nil, err
	}

	// [MAP_PEERS] Build lookup map and identify a default recipient (not sender) for the echo event
	peerMap := make(map[uuid.UUID]*model.Peer, len(peers))
	var recipient *model.Peer
	for i := range peers {
		p := &peers[i]
		peerMap[p.ID] = p
		if recipient == nil && p.ID != sID {
			recipient = p
		}
	}

	base := raw.ToDomain()
	if s, ok := peerMap[sID]; ok {
		base.From = *s
	}

	// [BUILD_EVENTS] Create events only for targets connected to this node
	events := make([]event.Eventer, 0, len(targets))
	for _, uid := range targets {
		p, ok := peerMap[uid]
		if !ok {
			continue
		}

		msg := *base
		// [RECIPIENT_LOGIC] Set 'To' field: recipient peer for sender, or self-peer for recipients
		msg.To = recipient
		if uid != sID {
			msg.To = p
		}
		events = append(events, event.NewMessageEvent(&msg, uid))
	}
	return events, nil
}

// [ON_THREAD_CREATED] Handles thread creation. Resolves ALL members for full UI state.
func (h *MessageHandler) OnThreadCreatedV1(ctx context.Context, raw *payload.ThreadCreatedV1) ([]event.Eventer, error) {
	memberIDs := toUUIDs(raw.Members)
	var targets []uuid.UUID

	// [FILTER_TARGETS] Only prepare events for locally connected users or if we are the leader
	if id, err := uuid.Parse(raw.Recipient.ID); err == nil {
		if h.dispatcher.IsLeader() || h.hub.Connected(id) {
			targets = []uuid.UUID{id}
		}
	} else {
		_, targets = h.filter("", memberIDs)
	}

	if len(targets) == 0 {
		return nil, nil
	}

	// [RESOLVE_ALL] For thread creation, we resolve ALL members to provide full chat context
	peers, err := h.enricher.Resolve(ctx, raw.DomainID, memberIDs...)
	if err != nil {
		return nil, err
	}

	base := raw.ToDomain()
	base.Members = peers

	// [BUILD_EVENTS] Dispatch the full thread object to each local target
	events := make([]event.Eventer, 0, len(targets))
	for _, uid := range targets {
		thread := *base
		events = append(events, event.NewThreadEvent(&thread, uid))
	}
	return events, nil
}

// [EXTRACT_PARTICIPANTS] Parses and deduplicates sender and recipient IDs.
func extractParticipants(raw *payload.MessageCreatedV1) (uuid.UUID, []uuid.UUID) {
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

// [FILTER] Decides which IDs should be processed by this node based on Hub connections.
func (h *MessageHandler) filter(senderIDStr string, all []uuid.UUID) (uuid.UUID, []uuid.UUID) {
	sID, _ := uuid.Parse(senderIDStr)
	if h.dispatcher.IsLeader() {
		return sID, all
	}
	res := make([]uuid.UUID, 0, len(all))
	for _, id := range all {
		if id == sID || h.hub.Connected(id) {
			res = append(res, id)
		}
	}
	return sID, res
}

// [TO_UUIDS] Helper for converting a slice of strings to UUIDs.
func toUUIDs(src []string) []uuid.UUID {
	res := make([]uuid.UUID, 0, len(src))
	for _, s := range src {
		if id, err := uuid.Parse(s); err == nil {
			res = append(res, id)
		}
	}
	return res
}
