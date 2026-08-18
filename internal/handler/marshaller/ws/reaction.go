package wsmarshaller

import (
	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// WSReaction is the WebSocket DTO of a reaction change on a message. It carries
// only the authoritative per-emoji state AFTER the change: the client replaces
// the message's whole reaction bar from Reactions. Reactions is always present
// (an empty array means the last reaction was removed).
//
// Each aggregate mirrors the im-thread message-history MessageReaction shape
// field-for-field, so the client can reuse the exact same rendering it already
// uses for history.
type WSReaction struct {
	MessageID string                `json:"message_id"`
	ThreadID  string                `json:"thread_id"`
	Reactions []WSReactionAggregate `json:"reactions"`
}

// WSReactionAggregate is one emoji's state on the message as seen by a single
// recipient — a copy of the im-thread history aggregate. ReactedByMe is derived
// server-side for this recipient. LastReactedAt is emitted as a string to match
// the im-thread int64/protojson wire representation.
type WSReactionAggregate struct {
	Emoji         string   `json:"emoji"`
	Count         int32    `json:"count"`
	ReactedByMe   bool     `json:"reacted_by_me"`
	ReactorIDs    []string `json:"reactor_ids,omitempty"`
	LastReactedAt int64    `json:"last_reacted_at,string"`
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
			ReactorIDs:    a.ReactorIDs,
			LastReactedAt: a.LastReactedAt,
		})
	}

	return &WSReaction{
		MessageID: m.ID.String(),
		ThreadID:  m.ThreadID.String(),
		Reactions: aggs,
	}
}
