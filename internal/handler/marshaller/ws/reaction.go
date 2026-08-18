package wsmarshaller

import (
	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// WSReaction is the WebSocket DTO of an emoji reaction on a message. Reactor is
// the enriched participant marshaled with the SAME helper as a message sender
// (mapPeer) — identical nested shape to a NewMessageEvent's `sender` and a
// Typing event's `from`, instead of the flat domain Peer fields.
type WSReaction struct {
	ID        string  `json:"id"`
	ThreadID  string  `json:"thread_id"`
	Reactor   *WSPeer `json:"reactor"`
	Emoji     string  `json:"emoji"`
	Removed   bool    `json:"removed"`
	ReactedAt int64   `json:"reacted_at"`
	SendID    string  `json:"send_id,omitempty"`

	// Reactions is the authoritative per-emoji aggregate on the message after
	// this change; the client replaces its reaction state from it.
	Reactions []WSReactionAggregate `json:"reactions,omitempty"`
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

	var aggs []WSReactionAggregate
	if len(m.Reactions) > 0 {
		aggs = make([]WSReactionAggregate, 0, len(m.Reactions))
		for _, a := range m.Reactions {
			aggs = append(aggs, WSReactionAggregate{
				Emoji:         a.Emoji,
				Count:         a.Count,
				ReactedByMe:   a.ReactedBy(vid),
				LastReactedAt: a.LastReactedAt,
			})
		}
	}

	return &WSReaction{
		ID:        m.ID.String(),
		ThreadID:  m.ThreadID.String(),
		Reactor:   mapPeer(&m.Reactor),
		Emoji:     m.Emoji,
		Removed:   m.Removed,
		ReactedAt: m.ReactedAt,
		SendID:    m.SendId,
		Reactions: aggs,
	}
}
