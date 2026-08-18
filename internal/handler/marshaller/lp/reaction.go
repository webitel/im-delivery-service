package lpmarshaller

import (
	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// lpReaction is the long-poll payload for a reaction change, seen by a single
// recipient. It mirrors the WebSocket shape: only the authoritative per-emoji
// state (Reactions) to replace the bar with, always present (empty array means
// no reactions left). Each aggregate is a copy of the im-thread history shape.
type lpReaction struct {
	MessageID string                `json:"message_id"`
	ThreadID  string                `json:"thread_id"`
	Reactions []lpReactionAggregate `json:"reactions"`
}

// lpReactionAggregate is one emoji's state on the message for this recipient,
// mirroring the im-thread history aggregate. LastReactedAt is emitted as a
// string to match the im-thread int64/protojson wire representation.
type lpReactionAggregate struct {
	Emoji         string   `json:"emoji"`
	Count         int32    `json:"count"`
	ReactedByMe   bool     `json:"reacted_by_me"`
	ReactorIDs    []string `json:"reactor_ids,omitempty"`
	LastReactedAt int64    `json:"last_reacted_at,string"`
}

// mapReaction builds the viewer-scoped long-poll reaction payload, resolving
// reacted_by_me per emoji from the aggregate's reactor ids.
func mapReaction(m *model.MessageReaction, viewer uuid.UUID) *lpReaction {
	vid := viewer.String()

	aggs := make([]lpReactionAggregate, 0, len(m.Reactions))
	for _, a := range m.Reactions {
		aggs = append(aggs, lpReactionAggregate{
			Emoji:         a.Emoji,
			Count:         a.Count,
			ReactedByMe:   a.ReactedBy(vid),
			ReactorIDs:    a.ReactorIDs,
			LastReactedAt: a.LastReactedAt,
		})
	}

	return &lpReaction{
		MessageID: m.ID.String(),
		ThreadID:  m.ThreadID.String(),
		Reactions: aggs,
	}
}
