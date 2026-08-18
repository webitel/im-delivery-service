package lpmarshaller

import (
	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// lpReaction is the long-poll payload for an emoji reaction change, seen by a
// single recipient. It mirrors the WebSocket shape: the reactor/emoji/removed
// fields describe the action, while Reactions is the authoritative per-emoji
// state to replace the bar with (always present, empty array means no
// reactions left).
type lpReaction struct {
	MessageID string                `json:"message_id"`
	ThreadID  string                `json:"thread_id"`
	Reactor   model.Peer            `json:"reactor"`
	Emoji     string                `json:"emoji"`
	Removed   bool                  `json:"removed"`
	ReactedAt int64                 `json:"reacted_at"`
	SendID    string                `json:"send_id,omitempty"`
	Reactions []lpReactionAggregate `json:"reactions"`
}

// lpReactionAggregate is one emoji's state on the message for this recipient.
type lpReactionAggregate struct {
	Emoji         string `json:"emoji"`
	Count         int32  `json:"count"`
	ReactedByMe   bool   `json:"reacted_by_me"`
	LastReactedAt int64  `json:"last_reacted_at"`
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
			LastReactedAt: a.LastReactedAt,
		})
	}

	return &lpReaction{
		MessageID: m.ID.String(),
		ThreadID:  m.ThreadID.String(),
		Reactor:   m.Reactor,
		Emoji:     m.Emoji,
		Removed:   m.Removed,
		ReactedAt: m.ReactedAt,
		SendID:    m.SendId,
		Reactions: aggs,
	}
}
