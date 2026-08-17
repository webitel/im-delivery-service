package wsmarshaller

import "github.com/webitel/im-delivery-service/internal/domain/model"

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
}

// mapReaction transforms the internal reaction domain into a WebSocket DTO.
func mapReaction(m *model.MessageReaction) *WSReaction {
	return &WSReaction{
		ID:        m.ID.String(),
		ThreadID:  m.ThreadID.String(),
		Reactor:   mapPeer(&m.Reactor),
		Emoji:     m.Emoji,
		Removed:   m.Removed,
		ReactedAt: m.ReactedAt,
		SendID:    m.SendId,
	}
}
