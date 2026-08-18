package wsmarshaller

import (
	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// WSReaction is the WebSocket DTO of an emoji reaction change on a message.
// The reactor/emoji/removed fields describe the single action that produced the
// event, while Reactions is the authoritative per-emoji state AFTER it — the
// client replaces the message's whole reaction bar from Reactions. Reactions is
// always present (an empty array means the last reaction was removed).
type WSReaction struct {
	MessageID string                `json:"message_id"`
	ThreadID  string                `json:"thread_id"`
	Reactor   *WSPeer               `json:"reactor"`
	Emoji     string                `json:"emoji"`
	Removed   bool                  `json:"removed"`
	ReactedAt int64                 `json:"reacted_at"`
	SendID    string                `json:"send_id,omitempty"`
	Reactions []WSReactionAggregate `json:"reactions"`
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
		Reactor:   mapPeer(&m.Reactor),
		Emoji:     m.Emoji,
		Removed:   m.Removed,
		ReactedAt: m.ReactedAt,
		SendID:    m.SendId,
		Reactions: aggs,
	}
}
