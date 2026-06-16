package amqp

import (
	"context"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
	"github.com/webitel/webitel-go-kit/pkg/semconv"
)

// OnThreadCreatedV1 handles thread creation. Resolves full UI state for targets.
func (h *MessageHandler) OnThreadCreatedV1(ctx context.Context, raw *payload.ThreadCreatedV1) ([]event.Eventer, error) {
	// 1. Extract Contact IDs for enrichment (names, types, etc.)
	memberIDs := toUUIDs(extractMemberIDs(raw.Members))
	var targets []uuid.UUID

	// Determine who needs to receive this event on this node
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

	// 2. Resolve basic peer metadata (Name, Type, Sub) from storage/cache
	peers, err := h.enricher.Resolve(ctx, raw.DomainID, memberIDs...)
	if err != nil {
		h.logger.Error("failed to enrich thread members", semconv.ErrorKey, err)
		return nil, err
	}

	// 3. Create a map for quick access during the overlay process
	peerMap := make(map[uuid.UUID]*model.Peer)
	for i := range peers {
		peerMap[peers[i].ID] = &peers[i]
	}

	// 4. [FIX] Overlay logic: Inject MemberID and Role from the RAW payload.
	// The enricher only knows 'who' they are, but the RAW payload knows 'what' they are in this thread.
	for _, rm := range raw.Members {
		cID, _ := uuid.Parse(rm.ContactID)
		if p, ok := peerMap[cID]; ok {
			p.MemberID = rm.MemberID // Sets the ID field in WSPeer
			p.Role = int32(rm.Role)  // Sets the Role string in WSPeer
		}
	}

	// 5. Build the domain thread template
	base := raw.ToDomain()

	// Re-assemble members list to ensure we use enriched & overlaid data
	base.Members = make([]model.Peer, 0, len(memberIDs))
	for _, id := range memberIDs {
		if p, ok := peerMap[id]; ok {
			base.Members = append(base.Members, *p)
		}
	}

	// 6. Fan-out events to targets
	events := make([]event.Eventer, 0, len(targets))
	for _, uid := range targets {
		// Create a shallow copy for each target event
		thread := *base
		events = append(events, event.NewThreadEvent(&thread, uid))
	}

	return events, nil
}

// extractMemberIDs extracts ContactID for the enrichment process.
func extractMemberIDs(members []payload.ThreadMember) []string {
	res := make([]string, 0, len(members))
	for _, m := range members {
		if m.ContactID != "" {
			res = append(res, m.ContactID)
		}
	}
	return res
}
