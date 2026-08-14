package wsmarshaller

import "github.com/webitel/im-delivery-service/internal/domain/model"

// WSTyping is the WebSocket DTO of an ephemeral "…is typing" indicator. From is
// the enriched typing participant marshaled with the SAME helper as a message
// sender (mapPeer) — identical shape to a NewMessageEvent's `sender`.
type WSTyping struct {
	ThreadID    string  `json:"thread_id"`
	TimeoutMs   int32   `json:"timeout_ms"`
	From        *WSPeer `json:"from"`
	PreviewText string  `json:"preview_text,omitempty"`
}

// mapTyping transforms the internal typing domain into a WebSocket DTO.
func mapTyping(t *model.Typing) *WSTyping {
	return &WSTyping{
		ThreadID:    t.ThreadID,
		TimeoutMs:   t.TimeoutMs,
		From:        mapPeer(&t.From),
		PreviewText: t.PreviewText,
	}
}
