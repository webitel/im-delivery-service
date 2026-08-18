package wsmarshaller

import (
	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// WSReaction is the WebSocket DTO of an emoji reaction change on a message. It
// separates two concerns so the client never has to reconcile a delta by hand:
//
//   - Reactions is the authoritative per-emoji state AFTER the change — the
//     client replaces the message's whole reaction bar from it. It is always
//     present (an empty array means the last reaction was removed).
//   - Actor describes the single action that produced this event — who reacted,
//     with what, and whether it was an add or a removal — for optional UI hints
//     (animations, "X reacted 🔥"). It is never needed to render the bar.
type WSReaction struct {
	MessageID string                `json:"message_id"`
	ThreadID  string                `json:"thread_id"`
	Reactions []WSReactionAggregate `json:"reactions"`
	Actor     WSReactionActor       `json:"actor"`
}

// WSReactionAggregate is one emoji's state on the message as seen by a single
// recipient. ReactedByMe is derived server-side from the aggregate's reactor
// ids, so the raw ids never reach the client.
type WSReactionAggregate struct {
	Emoji         string `json:"emoji"`
	Count         int32  `json:"count"`
	ReactedByMe   bool   `json:"reacted_by_me"`
	LastReactedAt int64  `json:"last_reacted_at"`
}

// WSReactionActor is the participant that triggered this event and what they
// did. Removed distinguishes an add from a removal; SendID echoes the actor's
// own request so their client can reconcile an optimistic update.
type WSReactionActor struct {
	Reactor   *WSPeer `json:"reactor"`
	Emoji     string  `json:"emoji"`
	Removed   bool    `json:"removed"`
	ReactedAt int64   `json:"reacted_at"`
	SendID    string  `json:"send_id,omitempty"`
}

// mapReaction transforms the internal reaction domain into a WebSocket DTO for
// the given viewer, resolving reacted_by_me per emoji.
func mapReaction(m *model.MessageReaction, viewer uuid.UUID) *WSReaction {
	vid := viewer.String()

	// Always emit a non-nil slice: an empty array is a meaningful "bar is now
	// empty" signal, not an absent field.
	aggs := make([]WSReactionAggregate, 0, len(m.Reactions))
	for _, a := range m.Reactions {
		aggs = append(aggs, WSReactionAggregate{
			Emoji:         a.Emoji,
			Count:         a.Count,
			ReactedByMe:   a.ReactedBy(vid),
			LastReactedAt: a.LastReactedAt,
		})
	}

	return &WSReaction{
		MessageID: m.ID.String(),
		ThreadID:  m.ThreadID.String(),
		Reactions: aggs,
		Actor: WSReactionActor{
			Reactor:   mapPeer(&m.Reactor),
			Emoji:     m.Emoji,
			Removed:   m.Removed,
			ReactedAt: m.ReactedAt,
			SendID:    m.SendId,
		},
	}
}
