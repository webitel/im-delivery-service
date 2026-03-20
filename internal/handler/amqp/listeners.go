package amqp

import (
	"context"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
)

// [ON_MESSAGE_CREATED] Handles message fan-out with Echo-aware logic.
func (h *MessageHandler) OnMessageCreatedV1(ctx context.Context, raw *payload.MessageCreatedV1) ([]event.Eventer, error) {
	senderID, participantIDs := h.extractParticipants(raw)

	// [FILTER] Identify targets this node is responsible for (local users + sender for echo).
	targets := h.computeLocalTargets(senderID, participantIDs)
	if len(targets) == 0 {
		return nil, nil
	}

	// [ENRICH] Resolve minimal metadata for all targets and the sender.
	peers, err := h.enricher.Resolve(ctx, raw.DomainID, h.getRequiredIDs(senderID, targets)...)
	if err != nil {
		return nil, err
	}

	// [INDEX] Map peers by ID for efficient lookup.
	peerMap := make(map[uuid.UUID]*model.Peer, len(peers))
	for i := range peers {
		peerMap[peers[i].ID] = &peers[i]
	}

	// [PREPARE] Build base domain model.
	template := raw.ToDomain()
	template.From = *peerMap[senderID]

	events := make([]event.Eventer, 0, len(targets))
	for _, targetID := range targets {
		isEcho := targetID == senderID

		// [IMMUTABILITY] Clone template to safely set per-recipient "To" field.
		msg := *template
		msg.To = h.resolveTargetPeer(targetID, senderID, peerMap)

		// [FACTORY] Create event with explicit Echo flag.
		events = append(events, event.NewMessageEvent(
			&msg,
			targetID,
			event.WithEcho[*model.Message](isEcho),
		))
	}

	return events, nil
}

// [COMPUTE_LOCAL_TARGETS] Filters participants based on leader status or active WebSocket connections.
func (h *MessageHandler) computeLocalTargets(senderID uuid.UUID, all []uuid.UUID) []uuid.UUID {
	if h.leader.IsLeader() {
		return all
	}
	res := make([]uuid.UUID, 0)
	for _, id := range all {
		// [SYNC_POLICY] Always include sender for Echo sync, or any user with active session.
		if id == senderID || h.hub.Connected(id) {
			res = append(res, id)
		}
	}
	return res
}

// [RESOLVE_TARGET_PEER] Logic to determine the 'To' peer for various message views.
func (h *MessageHandler) resolveTargetPeer(targetID, senderID uuid.UUID, peers map[uuid.UUID]*model.Peer) *model.Peer {
	if targetID != senderID {
		return peers[targetID]
	}
	// For Echo: return the first available recipient as a fallback peer.
	for id, p := range peers {
		if id != senderID {
			return p
		}
	}
	return nil
}

// [GET_REQUIRED_IDS] Utility to merge sender and targets into a resolution set.
func (h *MessageHandler) getRequiredIDs(sender uuid.UUID, targets []uuid.UUID) []uuid.UUID {
	res := append([]uuid.UUID{sender}, targets...)
	// Deduplicate is handled by the enrichment layer or simple slice manipulation.
	return res
}

// // [ON_MESSAGE_CREATED] Handles message fan-out with optimized enrichment.
// func (h *MessageHandler) OnMessageCreatedV1(ctx context.Context, raw *payload.MessageCreatedV1) ([]event.Eventer, error) {
// 	senderID, participants := extractParticipants(raw)

// 	// [FILTER] Identify local subscribers who should receive the event
// 	_, localTargets := h.filter(raw.From.ID, participants)
// 	if len(localTargets) == 0 {
// 		return nil, nil
// 	}

// 	// [RESOLVE_LOGIC] Determine minimal set of peers needed for metadata
// 	// If sender is local, we need everyone for the "To" field (echo sync)
// 	// Otherwise, we only need local recipients + the sender for the "From" field
// 	idsToResolve := localTargets
// 	isSenderPresent := slices.Contains(localTargets, senderID)

// 	if isSenderPresent {
// 		idsToResolve = participants
// 	} else if senderID != uuid.Nil {
// 		idsToResolve = append(slices.Clone(localTargets), senderID)
// 	}

// 	peers, err := h.enricher.Resolve(ctx, raw.DomainID, idsToResolve...)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// [INDEX] Map peers by ID and pick a default recipient for the sender's echo
// 	peerIndex := make(map[uuid.UUID]*model.Peer, len(peers))
// 	var fallbackRecipient *model.Peer
// 	for i := range peers {
// 		p := &peers[i]
// 		peerIndex[p.ID] = p
// 		if fallbackRecipient == nil && p.ID != senderID {
// 			fallbackRecipient = p
// 		}
// 	}

// 	// [PREPARE] Build base domain model
// 	msgTemplate := raw.ToDomain()
// 	if sender, ok := peerIndex[senderID]; ok {
// 		msgTemplate.From = *sender
// 	}

// 	// [DISPATCH] Build events for each local target
// 	events := make([]event.Eventer, 0, len(localTargets))
// 	for _, targetID := range localTargets {
// 		peer, ok := peerIndex[targetID]
// 		if !ok {
// 			continue
// 		}

// 		eventMsg := *msgTemplate
// 		eventMsg.To = fallbackRecipient // Default for sender's echo
// 		if targetID != senderID {
// 			eventMsg.To = peer // Self-info for recipients
// 		}

// 		events = append(events, event.NewMessageEvent(&eventMsg, targetID))
// 	}

// 	return events, nil
// }

// [ON_THREAD_CREATED] Handles thread creation. Resolves ALL members for full UI state.
func (h *MessageHandler) OnThreadCreatedV1(ctx context.Context, raw *payload.ThreadCreatedV1) ([]event.Eventer, error) {
	memberIDs := toUUIDs(raw.Members)
	var targets []uuid.UUID

	// [FILTER_TARGETS] Only prepare events for locally connected users or if we are the leader
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

// [FILTER] Decides which IDs should be processed by this node based on Hub connections.
func (h *MessageHandler) filter(senderIDStr string, all []uuid.UUID) (uuid.UUID, []uuid.UUID) {
	sID, _ := uuid.Parse(senderIDStr)
	if h.leader.IsLeader() {
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
